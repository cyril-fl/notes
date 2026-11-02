package readnote

import (
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/process/readitem"
)

// Handler shows a note: GET /api/notes/{id}.
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
