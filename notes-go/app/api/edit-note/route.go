package editnote

import (
	"net/http"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/patchnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	"github.com/uptrace/bun"
)

// Body is patchnote.Input in JSON: a field left out is kept.
type Body struct {
	Pinned   *bool     `json:"pinned" binding:"optional"`
	Readonly *bool     `json:"readonly" binding:"optional"`
	Content  *string   `json:"content" binding:"optional"`
	Tags     *[]string `json:"tags" binding:"optional"`
	// UpdatedAt is updated_at as read, or created_at for a note never edited: a 409 if the note changed since.
	UpdatedAt *time.Time `json:"updated_at" binding:"optional"`
	// Restore says the content is a version restored: the text it replaces is kept in the history.
	Restore bool `json:"restore,omitempty" binding:"optional"`
}

// Handler changes the fields given: PATCH /api/notes/{id}, through patchnote (shared with `edit note`).
// It answers the note as GET does, so the app needs no second request after a save.
// @Summary  Change the fields given
// @Tags     notes
// @Accept   json
// @Produce  json
// @Param    id path string true "uuid"
// @Param    body body Body true "body"
// @Success  200 {object} readitem.Result
// @Failure  400 {object} respond.Failure
// @Failure  404 {object} respond.Failure
// @Failure  409 {object} respond.Failure
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

		process := patchnote.Process(r.Context(), db, patchnote.Input{ID: id, Pinned: changes.Pinned, Readonly: changes.Readonly, Content: changes.Content, Tags: changes.Tags, UpdatedAt: changes.UpdatedAt, Restore: changes.Restore})
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		reply(w, process.GetResult())
	}
}

func reply(w http.ResponseWriter, note readitem.Result) {
	note.Tags = append([]string{}, note.Tags...)
	respond.JSON(w, http.StatusOK, note)
}
