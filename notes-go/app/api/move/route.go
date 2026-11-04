package move

import (
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/moveitems"
)

// Handler moves items into a folder, to the root without parent_id: POST /api/move.
// @Summary  Move items into a folder, or to the root without parent_id
// @Tags     items
// @Accept   json
// @Produce  json
// @Param    body body moveitems.Input true "body"
// @Success  200 {array} domain.Item
// @Failure  400 {object} respond.Failure
// @Failure  409 {object} respond.Failure
// @Failure  413 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/move [post]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input moveitems.Input
		if !respond.Decode(w, r, &input) || !respond.ValidIDs(w, input.ParentID) || !respond.ValidIDs(w, input.IDs...) {
			return
		}

		process := moveitems.Process(r.Context(), db, input)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.List(w, http.StatusOK, process.GetResult())
	}
}
