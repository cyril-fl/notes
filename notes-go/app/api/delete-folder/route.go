package deletefolder

import (
	"fmt"
	"net/http"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/deleteitems"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	"github.com/uptrace/bun"
)

// Handler moves the folder and its content to the trash: DELETE /api/folders/{id}.
// @Summary  Move a folder and its content to the trash
// @Tags     folders
// @Produce  json
// @Param    id path string true "uuid"
// @Success  204
// @Failure  404 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/folders/{id} [delete]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := respond.ID(w, r)
		if !ok {
			return
		}

		process := deleteitems.Process(r.Context(), db, domain.FOLDER, []string{id})
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
