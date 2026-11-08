package createtag

import (
	"net/http"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createtag"
	"github.com/uptrace/bun"
)

type Body struct {
	Values []string `json:"values"`
}

// Handler creates tags and returns the new ones, existing tags being skipped: POST /api/tags.
// @Summary  Create tags, skipping existing ones
// @Tags     tags
// @Accept   json
// @Produce  json
// @Param    body body Body true "body"
// @Success  201 {array} domain.Tag
// @Failure  400 {object} respond.Failure
// @Failure  413 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/tags [post]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input Body
		if !respond.Decode(w, r, &input) {
			return
		}

		process := createtag.Process(r.Context(), db, input.Values)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.List(w, http.StatusCreated, process.GetResult())
	}
}
