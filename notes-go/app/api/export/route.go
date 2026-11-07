package export

import (
	"archive/zip"
	"log"
	"net/http"
	"time"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/exportnotes"
)

// Handler gives every live note as a zip of markdown files, one directory per folder: GET /api/export.
// @Summary  Export the notes as markdown files
// @Tags     transfer
// @Produce  application/zip
// @Success  200 {file} binary
// @Failure  500 {object} respond.Failure
// @Router   /api/export [get]
func Handler(db bun.IDB) http.HandlerFunc {
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
				_, err = entry.Write([]byte(file.Content))
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
