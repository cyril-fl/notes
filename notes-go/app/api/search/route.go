package search

import (
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/searchnotes"
)

// Handler finds notes by title or content: GET /api/search?q=.
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
