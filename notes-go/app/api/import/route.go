package importzip

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/importnotes"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/silo"
	"github.com/uptrace/bun"
)

const (
	// maxZip limits the upload, maxContent what it unpacks to: text and images, these are far beyond real notes.
	maxZip     = 200 << 20
	maxContent = 400 << 20
)

// Handler adds the .md files of a zip as notes, its directories as folders, into a folder (the root without
// parent), all or nothing: POST /api/import?parent= with the zip as body, or a multipart form of files named
// by their relative path (dir/note.md), the loose files or the folder picked in the browser.
// @Summary  Import markdown files from a zip or a form of files
// @Tags     transfer
// @Accept   application/zip
// @Accept   multipart/form-data
// @Produce  json
// @Param    parent query string false "folder uuid, the root when left out"
// @Param    body body []byte true "zip of .md files and their images, or a form of files named by their path"
// @Success  200 {object} importnotes.Result
// @Failure  400 {object} respond.Failure
// @Failure  409 {object} respond.Failure
// @Failure  413 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/import [post]
func Handler(db bun.IDB, store silo.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parent := r.URL.Query().Get("parent")
		if !respond.ValidIDs(w, parent) {
			return
		}

		// A big zip from the phone outlasts the server's 30 s: the write deadline also runs from the headers.
		control := http.NewResponseController(w)
		_ = control.SetReadDeadline(time.Time{})
		_ = control.SetWriteDeadline(time.Time{})

		r.Body = http.MaxBytesReader(w, r.Body, maxZip)
		dir, clean, err := read(r)
		if clean != nil {
			defer clean()
		}
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			respond.JSON(w, http.StatusRequestEntityTooLarge, respond.Failure{Error: "request body too large"})
			return
		}
		if errors.Is(err, errBadBody) || errors.Is(err, errDuplicate) {
			respond.JSON(w, http.StatusBadRequest, respond.Failure{Error: err.Error()})
			return
		}
		if errors.Is(err, errTooLarge) {
			respond.JSON(w, http.StatusRequestEntityTooLarge, respond.Failure{Error: err.Error()})
			return
		}
		if err != nil {
			respond.Error(w, err)
			return
		}

		process := importnotes.Process(r.Context(), db, store, importnotes.Input{Dir: dir, ParentID: parent})
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.JSON(w, http.StatusOK, process.GetResult())
	}
}

var (
	errBadBody  = errors.New("not a zip file nor a form of files")
	errTooLarge = errors.New("zip content too large")
	// Loose files have no folder in their path: two of the same name would keep the last one only (#69).
	errDuplicate = errors.New("two files have the same path")
)

// read gives the files of the body: a zip, or a multipart form. Both land in a temporary file, never in memory:
// the server shares its machine with Postgres and Silo, and a 200 MB import held in RAM, plus the sealed copy
// of every file, had the OOM killer end it halfway (#119). clean removes what was written, nil only when
// nothing was.
func read(r *http.Request) (fs.FS, func(), error) {
	if form, err := r.MultipartReader(); err == nil {
		return readForm(form)
	}

	file, err := os.CreateTemp("", "import-*.zip")
	if err != nil {
		return nil, nil, err
	}
	clean := func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}
	size, err := io.Copy(file, r.Body)
	if err != nil {
		return nil, clean, err // Still a *http.MaxBytesError past the limit.
	}
	archive, err := zip.NewReader(file, size)
	if err != nil {
		return nil, clean, errBadBody
	}
	var content uint64
	for _, entry := range archive.File {
		content += entry.UncompressedSize64
	}
	if content > maxContent {
		return nil, clean, errTooLarge
	}
	return archive, clean, nil
}

// readForm writes every file of the form at its path under a temporary directory, its directories made from it.
// The body's limit bounds the content: maxZip is below maxContent.
func readForm(form *multipart.Reader) (fs.FS, func(), error) {
	root, err := os.MkdirTemp("", "import-*")
	if err != nil {
		return nil, nil, err
	}
	clean := func() { _ = os.RemoveAll(root) }
	for {
		part, err := form.NextPart()
		// Not errors.Is: a cut body ends on a wrapped io.EOF, the end of the form on a bare one.
		if err == io.EOF {
			return os.DirFS(root), clean, nil
		}
		if err != nil {
			return nil, clean, fmt.Errorf("%w: %w", errBadBody, err) // Still a *http.MaxBytesError past the limit.
		}
		// Not part.FileName(): it keeps the base name only, the path holds the folders.
		_, params, _ := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
		name := path.Clean(params["filename"])
		if params["filename"] == "" {
			continue // A field, not a file.
		}
		// fs.ValidPath refuses an absolute path and a .. that would write outside root.
		if !fs.ValidPath(name) || name == "." {
			return nil, clean, errBadBody
		}
		if err := writeFile(root, name, part); err != nil {
			return nil, clean, err
		}
	}
}

// writeFile writes part at name under root. O_EXCL answers errDuplicate: loose files have no folder in their
// path, and two of the same name would keep the last one only (#69).
func writeFile(root, name string, part io.Reader) error {
	full := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
		return fmt.Errorf("%w: %w", errBadBody, err) // A file where a directory is needed, mostly.
	}
	file, err := os.OpenFile(full, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("%w: %s", errDuplicate, name)
	}
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := io.Copy(file, part); err != nil {
		return fmt.Errorf("%w: %w", errBadBody, err)
	}
	return nil
}
