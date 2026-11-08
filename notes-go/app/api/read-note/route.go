package readnote

import (
	"net/http"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	"github.com/uptrace/bun"
)

// Handler shows a note: GET /api/notes/{id}.
// @Summary  Read a note
// @Tags     notes
// @Produce  json
// @Param    id path string true "uuid"
// @Success  200 {object} readitem.Result
// @Failure  404 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/notes/{id} [get]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := respond.ID(w, r)
		if !ok {
			return
		}

		process := readitem.Process(r.Context(), db, domain.NOTE, id)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		note := process.GetResult()
		note.Tags = append([]string{}, note.Tags...)
		respond.JSON(w, http.StatusOK, note)
	}
}
