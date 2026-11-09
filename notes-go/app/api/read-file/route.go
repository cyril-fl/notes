package readfile

import (
	"io"
	"log"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readfile"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/silo"
	"github.com/uptrace/bun"
)

// Handler sends an image or an attachment of the account: GET /api/files/{id}. Through the API, never Silo itself: the session
// guards it like the notes.
// @Summary  Read an image
// @Tags     files
// @Produce  image/png,image/jpeg,image/gif,image/webp,application/pdf,application/zip,application/ogg,audio/mpeg,audio/wave,video/mp4,video/webm
// @Param    id path string true "uuid"
// @Success  200 {file} binary
// @Failure  404 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/files/{id} [get]
func Handler(db bun.IDB, store silo.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := respond.ID(w, r)
		if !ok {
			return
		}

		process := readfile.Process(r.Context(), db, store, id)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		body := process.GetBody()
		defer body.Close()

		file := process.GetResult()
		w.Header().Set("Content-Type", file.ContentType)
		w.Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// An attachment downloads under its name: the note stays open, nothing of it renders on the API's origin.
		if !strings.HasPrefix(file.ContentType, "image/") {
			w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": string(file.Name)}))
		}
		// A file never changes: a new image gets a new id.
		w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
		// A full-size photo on a slow mobile link outlasts the server's 30 s, like its upload.
		_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
		if _, err := io.Copy(w, body); err != nil {
			// The status is sent: the image breaks off, the browser shows it failed.
			log.Printf("read file %s: %v", id, err)
		}
	}
}
