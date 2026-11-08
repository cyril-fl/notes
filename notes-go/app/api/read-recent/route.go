package readrecent

import (
	"net/http"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listrecent"
	"github.com/uptrace/bun"
)

// Handler lists the last changed notes: GET /api/recent.
// @Summary  List the last changed notes
// @Tags     search
// @Produce  json
// @Success  200 {array} listrecent.Recent
// @Failure  500 {object} respond.Failure
// @Router   /api/recent [get]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		process := listrecent.Process(r.Context(), db)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.List(w, http.StatusOK, process.GetResult())
	}
}
