package deletetag

import (
	"fmt"
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/deletetag"
	"main.go/app/shared/services/process/readtag"
)

// Handler deletes a tag, not the notes carrying it: DELETE /api/tags/{value}.
// @Summary  Delete a tag and erase it from the text of its notes
// @Tags     tags
// @Produce  json
// @Param    value path string true "tag"
// @Success  204
// @Failure  404 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/tags/{value} [delete]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		value := r.PathValue("value")
		process := deletetag.Process(r.Context(), db, []string{value})
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		// deletetag skips what it does not find: for one tag, that is a 404.
		if len(process.GetResult()) == 0 {
			respond.Error(w, fmt.Errorf("%w: %q", readtag.ErrNotFound, value))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
