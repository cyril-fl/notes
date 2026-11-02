package deletefolder

import (
	"fmt"
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/process/deleteitems"
	"main.go/app/shared/services/process/readitem"
)

// Handler moves the folder and its content to the trash: DELETE /api/folders/{id}.
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
