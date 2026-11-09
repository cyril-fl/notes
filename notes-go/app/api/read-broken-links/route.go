package readbrokenlinks

import (
	"net/http"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listbrokenlinks"
	"github.com/uptrace/bun"
)

// Handler lists the links of the live notes that point to no note, or to a title several notes have: GET /api/links/broken.
// @Summary  List the broken and ambiguous links
// @Tags     graph
// @Produce  json
// @Success  200 {array} listbrokenlinks.Link
// @Failure  500 {object} respond.Failure
// @Router   /api/links/broken [get]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		process := listbrokenlinks.Process(r.Context(), db)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.List(w, http.StatusOK, process.GetResult())
	}
}
