package importzip

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"mime/multipart"
	"net/http"
	"path"
	"testing/fstest"
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
		dir, err := read(r)
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			respond.JSON(w, http.StatusRequestEntityTooLarge, respond.Failure{Error: "request body too large"})
			return
		}
		if errors.Is(err, errBadBody) {
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
)

// read gives the files of the body: a zip, or a multipart form.
func read(r *http.Request) (fs.FS, error) {
	if form, err := r.MultipartReader(); err == nil {
		return readForm(form)
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, errBadBody
	}
	var size uint64
	for _, file := range archive.File {
		size += file.UncompressedSize64
	}
	if size > maxContent {
		return nil, errTooLarge
	}
	return archive, nil
}

// readForm keeps every file of the form at its path, its directories made from it. The body's limit bounds
// the content: maxZip is below maxContent.
func readForm(form *multipart.Reader) (fs.FS, error) {
	dir := fstest.MapFS{}
	for {
		part, err := form.NextPart()
		// Not errors.Is: a cut body ends on a wrapped io.EOF, the end of the form on a bare one.
		if err == io.EOF {
			return dir, nil
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errBadBody, err) // Still a *http.MaxBytesError past the limit.
		}
		// Not part.FileName(): it keeps the base name only, the path holds the folders.
		_, params, _ := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
		name := path.Clean(params["filename"])
		if params["filename"] == "" {
			continue // A field, not a file.
		}
		if !fs.ValidPath(name) || name == "." {
			return nil, errBadBody
		}
		data, err := io.ReadAll(part)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errBadBody, err)
		}
		dir[name] = &fstest.MapFile{Data: data}
	}
}
