package editnote

import (
	"net/http"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/hashtag"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/editnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/pinnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	"github.com/uptrace/bun"
)

// Body has pointers to tell a field left out from an empty one: editnote replaces everything,
// so a request without tags must keep them, not clear them.
type Body struct {
	Pinned  *bool     `json:"pinned" binding:"optional"`
	Content *string   `json:"content" binding:"optional"`
	Tags    *[]string `json:"tags" binding:"optional"`
}

// Handler changes the fields given: PATCH /api/notes/{id}. The #tags of the content are the tags of the note,
// its first heading (or first words) its title. Pinning alone leaves the note as it is, updated_at included.
// shortcut: reads then writes outside a transaction, a concurrent edit can be lost, fine for one user; use a tx if that changes.
// @Summary  Change the fields given
// @Tags     notes
// @Accept   json
// @Produce  json
// @Param    id path string true "uuid"
// @Param    body body Body true "body"
// @Success  200 {object} domain.Item
// @Failure  400 {object} respond.Failure
// @Failure  404 {object} respond.Failure
// @Failure  413 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/notes/{id} [patch]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := respond.ID(w, r)
		if !ok {
			return
		}

		var changes Body
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
		if changes.Pinned != nil {
			pin := pinnote.Process(r.Context(), db, domain.NOTE, id, *changes.Pinned)
			pin.Run()
			if pin.HasErr() {
				respond.Error(w, pin.GetErr())
				return
			}
		}
		if changes.Content == nil && changes.Tags == nil {
			respond.JSON(w, http.StatusOK, note.Item)
			return
		}

		input := editnote.Input{ID: id, Content: note.Content, Tags: note.Tags}
		if changes.Content != nil {
			input.Content = *changes.Content
		}
		// The text decides, unless the tags were asked for: they then rewrite it.
		input.Tags = hashtag.Parse(input.Content)
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
