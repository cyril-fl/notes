package export

import (
	"archive/zip"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/exportnotes"
	"main.go/app/shared/services/silo"
)

// Handler gives every live note as a zip of markdown files, one directory per folder, the images in _files:
// GET /api/export.
// @Summary  Export the notes as markdown files
// @Tags     transfer
// @Produce  application/zip
// @Success  200 {file} binary
// @Failure  500 {object} respond.Failure
// @Router   /api/export [get]
func Handler(db bun.IDB, store silo.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		process := exportnotes.Process(r.Context(), db, "")
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", `attachment; filename="notes-`+time.Now().Format("2006-01-02")+`.zip"`)
		archive := zip.NewWriter(w)
		for _, file := range process.GetResult() {
			name := file.Path
			if file.Dir {
				name += "/" // Keeps the empty folders.
			}
			entry, err := archive.Create(name)
			if err == nil {
				err = write(r, store, entry, file)
			}
			if err != nil {
				// The status is sent: the download breaks off, the browser shows it failed.
				log.Printf("write export: %v", err)
				return
			}
		}
		if err := archive.Close(); err != nil {
			log.Printf("write export: %v", err)
		}
	}
}

// write copies a note's text, or an image's bytes from Silo.
func write(r *http.Request, store silo.Store, entry io.Writer, file exportnotes.File) error {
	if file.Key == "" {
		_, err := io.WriteString(entry, file.Content)
		return err
	}
	body, err := store.Get(r.Context(), file.Key)
	if err != nil {
		return err
	}
	defer body.Close()
	_, err = io.Copy(entry, body)
	return err
}
