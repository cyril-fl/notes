package deletetrash

import (
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/emptytrash"
)

// Handler empties the trash for good: DELETE /api/trash.
// @Summary  Empty the trash for good
// @Tags     trash
// @Produce  json
// @Success  204
// @Failure  500 {object} respond.Failure
// @Router   /api/trash [delete]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		process := emptytrash.Process(r.Context(), db)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// Item deletes one removal of the trash for good, with what went with it: DELETE /api/trash/{id}.
// @Summary  Delete one removal of the trash for good
// @Tags     trash
// @Produce  json
// @Param    id path string true "uuid of an item listed by GET /api/trash"
// @Success  204
// @Failure  404 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/trash/{id} [delete]
func Item(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := respond.ID(w, r)
		if !ok {
			return
		}

		process := emptytrash.Process(r.Context(), db, id)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
