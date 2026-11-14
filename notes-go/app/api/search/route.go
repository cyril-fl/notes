package search

import (
	"net/http"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/searchnotes"
	"github.com/uptrace/bun"
)

// Handler finds notes by title or content: GET /api/search?q=&limit=&offset=, X-Total-Count counting every match.
// @Summary  Find notes by title or content
// @Tags     search
// @Produce  json
// @Param    q      query string  true  "text to find"
// @Param    limit  query integer false "matches in the page, all when left out or 0"
// @Param    offset query integer false "matches skipped before the page"
// @Success  200 {array} searchnotes.Match
// @Header   200 {integer} X-Total-Count "every match, not only the page"
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
		matches, ok := respond.Page(w, r, process.GetResult())
		if !ok {
			return
		}
		respond.List(w, http.StatusOK, matches)
	}
}
