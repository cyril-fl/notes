package createfile

import (
	"errors"
	"io"
	"net/http"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createfile"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/silo"
	"github.com/uptrace/bun"
)

// maxFile is far beyond a photo of a phone.
const maxFile = 20 << 20

// Handler stores an image, the body, under its name: POST /api/files?name=. The note cites it as /api/files/{id}.
// @Summary  Upload an image
// @Tags     files
// @Accept   image/png,image/jpeg,image/gif,image/webp
// @Produce  json
// @Param    name query string false "file name"
// @Param    body body []byte true "image"
// @Success  201 {object} domain.File
// @Failure  400 {object} respond.Failure
// @Failure  413 {object} respond.Failure
// @Failure  415 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/files [post]
func Handler(db bun.IDB, store silo.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxFile))
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			respond.JSON(w, http.StatusRequestEntityTooLarge, respond.Failure{Error: "request body too large"})
			return
		}
		if err != nil {
			respond.Error(w, err)
			return
		}

		process := createfile.Process(r.Context(), db, store, createfile.Input{Name: r.URL.Query().Get("name"), Data: data})
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.JSON(w, http.StatusCreated, process.GetResult())
	}
}
