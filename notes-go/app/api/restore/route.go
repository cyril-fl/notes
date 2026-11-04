package restore

import (
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/restoreitems"
)

type Body struct {
	IDs []string `json:"ids"`
}

// Handler takes items out of the trash, with everything removed at the same time: POST /api/restore.
// @Summary  Take items out of the trash
// @Tags     trash
// @Accept   json
// @Produce  json
// @Param    body body Body true "body"
// @Success  200 {array} domain.Item
// @Failure  400 {object} respond.Failure
// @Failure  409 {object} respond.Failure
// @Failure  413 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/restore [post]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input Body
		if !respond.Decode(w, r, &input) || !respond.ValidIDs(w, input.IDs...) {
			return
		}

		process := restoreitems.Process(r.Context(), db, input.IDs)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.List(w, http.StatusOK, process.GetResult())
	}
}
