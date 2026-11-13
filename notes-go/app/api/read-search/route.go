// Package readsearch is the API route running a saved search.
package readsearch

import (
	"net/http"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readsavedsearch"
	"github.com/uptrace/bun"
)

// Handler runs a saved search and lists the notes it matches: GET /api/searches/{id}.
// @Summary  Run a saved search
// @Tags     searches
// @Produce  json
// @Param    id path string true "saved search id"
// @Success  200 {array} runsavedsearch.Match
// @Failure  404 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/searches/{id} [get]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		process := readsavedsearch.Process(r.Context(), db, r.PathValue("id"))
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.List(w, http.StatusOK, process.GetResult())
	}
}
