package readversions

import (
	"net/http"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listversions"
	"github.com/uptrace/bun"
)

// Handler lists the older texts of a note: GET /api/notes/{id}/versions.
// @Summary  List the versions of a note
// @Tags     notes
// @Produce  json
// @Param    id path string true "uuid"
// @Success  200 {array} domain.NoteVersion
// @Failure  404 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/notes/{id}/versions [get]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := respond.ID(w, r)
		if !ok {
			return
		}

		process := listversions.Process(r.Context(), db, id)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.List(w, http.StatusOK, process.GetResult())
	}
}
