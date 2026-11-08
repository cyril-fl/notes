package readfile

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readfile"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/silo"
	"github.com/uptrace/bun"
)

// Handler sends an image of the account: GET /api/files/{id}. Through the API, never Silo itself: the session
// guards it like the notes.
// @Summary  Read an image
// @Tags     files
// @Produce  image/png,image/jpeg,image/gif,image/webp
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
		// A file never changes: a new image gets a new id.
		w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
		if _, err := io.Copy(w, body); err != nil {
			// The status is sent: the image breaks off, the browser shows it failed.
			log.Printf("read file %s: %v", id, err)
		}
	}
}
