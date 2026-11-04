package readfinder

import (
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/listitems"
)

// Handler lists the direct children of a folder, of the root without parent: GET /api/items?parent=.
// @Summary  List the direct children of a folder, or of the root
// @Tags     items
// @Produce  json
// @Param    parent query string false "folder uuid, the root when left out"
// @Success  200 {array} domain.Item
// @Failure  400 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/items [get]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parent := r.URL.Query().Get("parent")
		if !respond.ValidIDs(w, parent) {
			return
		}

		process := listitems.Process(r.Context(), db, parent)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.List(w, http.StatusOK, process.GetResult())
	}
}
