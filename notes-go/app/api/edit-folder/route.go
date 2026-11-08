package editfolder

import (
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

		if changes.Pinned != nil {
			pin := pinnote.Process(r.Context(), db, domain.FOLDER, id, *changes.Pinned)
			pin.Run()
			if pin.HasErr() {
				respond.Error(w, pin.GetErr())
				return
			}
		}
		if changes.Title == nil {
			current := readitem.Process(r.Context(), db, domain.FOLDER, id)
			current.Run()
			if current.HasErr() {
				respond.Error(w, current.GetErr())
				return
			}
			respond.JSON(w, http.StatusOK, current.GetResult().Item)
			return
		}

		process := editfolder.Process(r.Context(), db, editfolder.Input{ID: id, Title: *changes.Title})
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.JSON(w, http.StatusOK, process.GetResult())
	}
}
