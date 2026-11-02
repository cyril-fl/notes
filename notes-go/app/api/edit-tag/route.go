package edittag

import (
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/edittag"
)

type body struct {
	Value string `json:"value"`
}

// Handler renames a tag: PATCH /api/tags/{value}.
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input body
		if !respond.Decode(w, r, &input) {
			return
		}

		process := edittag.Process(r.Context(), db, r.PathValue("value"), input.Value)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.JSON(w, http.StatusOK, process.GetResult())
	}
}
