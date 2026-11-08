package search

import (
	"net/http"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/searchnotes"
	"github.com/uptrace/bun"
)

// Handler finds notes by title or content: GET /api/search?q=.
// @Summary  Find notes by title or content
// @Tags     search
// @Produce  json
// @Param    q query string true "text to find"
// @Success  200 {array} searchnotes.Match
// @Failure  400 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/search [get]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		process := searchnotes.Process(r.Context(), db, r.URL.Query().Get("q"))
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.List(w, http.StatusOK, process.GetResult())
	}
}
