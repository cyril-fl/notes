package editfolder

import (
	"context"
	"net/http"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/editfolder"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/pinnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	"github.com/uptrace/bun"
)

// Body has pointers to tell a field left out: {"pinned": true} alone keeps the title.
type Body struct {
	Title  *string `json:"title" binding:"optional"`
	Pinned *bool   `json:"pinned" binding:"optional"`
}

// Handler renames a folder or makes it a favorite: PATCH /api/folders/{id}.
// @Summary  Rename a folder, or make it a favorite
// @Tags     folders
// @Accept   json
// @Produce  json
// @Param    id path string true "uuid"
// @Param    body body Body true "body"
// @Success  200 {object} domain.Item
// @Failure  400 {object} respond.Failure
// @Failure  409 {object} respond.Failure
// @Failure  404 {object} respond.Failure
// @Failure  413 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/folders/{id} [patch]
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

		var folder domain.Item
		// The ctx of the request, not the one RunInTx hands: a bun.Tx's own ctx may act for another account.
		err := db.RunInTx(r.Context(), nil, func(_ context.Context, tx bun.Tx) error {
			var err error
			folder, err = patch(r.Context(), tx, id, changes)
			return err
		})
		if err != nil {
			respond.Error(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, folder)
	}
}

// patch pins then renames in tx, so a failed rename leaves the favorite as it was.
func patch(ctx context.Context, tx bun.IDB, id string, changes Body) (domain.Item, error) {
	if changes.Pinned != nil {
		pin := pinnote.Process(ctx, tx, domain.FOLDER, id, *changes.Pinned)
		pin.Run()
		if pin.HasErr() {
			return domain.Item{}, pin.GetErr()
		}
	}
	if changes.Title == nil {
		current := readitem.Process(ctx, tx, domain.FOLDER, id)
		current.Run()
		if current.HasErr() {
			return domain.Item{}, current.GetErr()
		}
		return current.GetResult().Item, nil
	}

	process := editfolder.Process(ctx, tx, editfolder.Input{ID: id, Title: *changes.Title})
	process.Run()
	if process.HasErr() {
		return domain.Item{}, process.GetErr()
	}
	return process.GetResult(), nil
}
