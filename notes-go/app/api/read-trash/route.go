package readtrash

import (
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/listtrash"
)

// Handler lists what was removed: GET /api/trash.
// @Summary  List what was removed
// @Tags     trash
// @Produce  json
// @Success  200 {array} listtrash.Trashed
// @Failure  500 {object} respond.Failure
// @Router   /api/trash [get]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		process := listtrash.Process(r.Context(), db)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.List(w, http.StatusOK, process.GetResult())
	}
}
