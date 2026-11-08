package createnote

import (
	"net/http"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/uptrace/bun"
)

// Handler creates a note: POST /api/notes.
// @Summary  Create a note
// @Tags     notes
// @Accept   json
// @Produce  json
// @Param    body body createnote.Input true "body"
// @Success  201 {object} domain.Item
// @Failure  400 {object} respond.Failure
// @Failure  413 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/notes [post]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input createnote.Input
		if !respond.Decode(w, r, &input) || !respond.ValidIDs(w, input.ParentID) {
			return
		}

		process := createnote.Process(r.Context(), db, input)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.JSON(w, http.StatusCreated, process.GetResult())
	}
}
