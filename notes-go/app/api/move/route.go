package move

import (
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/moveitems"
)

// Handler moves items into a folder, to the root without parent_id: POST /api/move.
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input moveitems.Input
		if !respond.Decode(w, r, &input) || !respond.ValidIDs(w, input.ParentID) || !respond.ValidIDs(w, input.IDs...) {
			return
		}

		process := moveitems.Process(r.Context(), db, input)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.List(w, http.StatusOK, process.GetResult())
	}
}
