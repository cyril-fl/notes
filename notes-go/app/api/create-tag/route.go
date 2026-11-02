package createtag

import (
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/createtag"
)

type body struct {
	Values []string `json:"values"`
}

// Handler creates tags and returns the new ones, existing tags being skipped: POST /api/tags.
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input body
		if !respond.Decode(w, r, &input) {
			return
		}

		process := createtag.Process(r.Context(), db, input.Values)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.List(w, http.StatusCreated, process.GetResult())
	}
}
