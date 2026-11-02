package readfinder

import (
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/listitems"
)

// Handler lists the direct children of a folder, of the root without parent: GET /api/items?parent=.
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
