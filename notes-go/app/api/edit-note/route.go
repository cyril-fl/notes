package editnote

import (
	"context"
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
// It answers the note as GET does, so the app needs no second request after a save.
// One transaction: a failed edit undoes the pin of the same request (#39).
// shortcut: the read is not locked, a concurrent edit can still be lost, fine for one user; SELECT … FOR UPDATE if that changes.
// @Summary  Change the fields given
// @Tags     notes
// @Accept   json
// @Produce  json
// @Param    id path string true "uuid"
// @Param    body body Body true "body"
// @Success  200 {object} readitem.Result
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

		var note readitem.Result
		// The ctx of the request, not the one RunInTx hands: a bun.Tx's own ctx may act for another account.
		err := db.RunInTx(r.Context(), nil, func(_ context.Context, tx bun.Tx) error {
			var err error
			note, err = patch(r.Context(), tx, id, changes)
			return err
		})
		if err != nil {
			respond.Error(w, err)
			return
		}
		reply(w, note)
	}
}

// patch pins then edits in tx, so a failed edit leaves the pin as it was.
func patch(ctx context.Context, tx bun.IDB, id string, changes Body) (readitem.Result, error) {
	current := readitem.Process(ctx, tx, domain.NOTE, id)
	current.Run()
	if current.HasErr() {
		return readitem.Result{}, current.GetErr()
	}
	note := current.GetResult()
	if changes.Pinned != nil {
		pin := pinnote.Process(ctx, tx, domain.NOTE, id, *changes.Pinned)
		pin.Run()
		if pin.HasErr() {
			return readitem.Result{}, pin.GetErr()
		}
	}
	if changes.Content == nil && changes.Tags == nil {
		if changes.Pinned != nil {
			note.Pinned = *changes.Pinned
		}
		return note, nil
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

	process := editnote.Process(ctx, tx, input)
	process.Run()
	if process.HasErr() {
		return readitem.Result{}, process.GetErr()
	}
	// The saved text is read back: editnote has turned its [[links]] into note ids.
	saved := readitem.Process(ctx, tx, domain.NOTE, id)
	saved.Run()
	if saved.HasErr() {
		return readitem.Result{}, saved.GetErr()
	}
	return saved.GetResult(), nil
}

func reply(w http.ResponseWriter, note readitem.Result) {
	note.Tags = append([]string{}, note.Tags...)
	respond.JSON(w, http.StatusOK, note)
}
