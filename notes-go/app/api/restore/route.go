package restore

import (
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/restoreitems"
)

type body struct {
	IDs []string `json:"ids"`
}

// Handler takes items out of the trash, with everything removed at the same time: POST /api/restore.
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input body
		if !respond.Decode(w, r, &input) || !respond.ValidIDs(w, input.IDs...) {
			return
		}

		process := restoreitems.Process(r.Context(), db, input.IDs)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.List(w, http.StatusOK, process.GetResult())
	}
}
