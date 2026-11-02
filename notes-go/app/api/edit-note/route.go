package editnote

import (
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/process/editnote"
	"main.go/app/shared/services/process/readitem"
)

// body has pointers to tell a field left out from an empty one: editnote replaces everything,
// so a request without tags must keep them, not clear them.
type body struct {
	Title   *string   `json:"title"`
	Content *string   `json:"content"`
	Tags    *[]string `json:"tags"`
}

// Handler changes the fields given: PATCH /api/notes/{id}.
// shortcut: reads then writes outside a transaction, a concurrent edit can be lost, fine for one user; use a tx if that changes.
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := respond.ID(w, r)
		if !ok {
			return
		}

		var changes body
		if !respond.Decode(w, r, &changes) {
			return
		}

		current := readitem.Process(r.Context(), db, domain.NOTE, id)
		current.Run()
		if current.HasErr() {
			respond.Error(w, current.GetErr())
			return
		}
		note := current.GetResult()
		input := editnote.Input{ID: id, Title: note.Item.Title, Content: note.Content, Tags: note.Tags}
		if changes.Title != nil {
			input.Title = *changes.Title
		}
		if changes.Content != nil {
			input.Content = *changes.Content
		}
		if changes.Tags != nil {
			input.Tags = *changes.Tags
		}

		process := editnote.Process(r.Context(), db, input)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.JSON(w, http.StatusOK, process.GetResult())
	}
}
