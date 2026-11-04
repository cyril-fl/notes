package deletenote

import (
	"fmt"
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/process/deleteitems"
	"main.go/app/shared/services/process/readitem"
)

// Handler moves the note to the trash: DELETE /api/notes/{id}.
// @Summary  Move a note to the trash
// @Tags     notes
// @Produce  json
// @Param    id path string true "uuid"
// @Success  204
// @Failure  404 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/notes/{id} [delete]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := respond.ID(w, r)
		if !ok {
			return
		}

		process := deleteitems.Process(r.Context(), db, domain.NOTE, []string{id})
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		// deleteitems skips what it does not find: for one item, that is a 404.
		if len(process.GetResult()) == 0 {
			respond.Error(w, fmt.Errorf("%w: %q", readitem.ErrNotFound, id))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
