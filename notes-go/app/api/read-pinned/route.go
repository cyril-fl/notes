package readpinned

import (
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/listpinned"
)

// Handler lists the pinned notes: GET /api/pinned.
// @Summary  List the pinned notes
// @Tags     search
// @Produce  json
// @Success  200 {array} listpinned.Pinned
// @Failure  500 {object} respond.Failure
// @Router   /api/pinned [get]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		process := listpinned.Process(r.Context(), db)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.List(w, http.StatusOK, process.GetResult())
	}
}
