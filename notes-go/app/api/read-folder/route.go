package readfolder

import (
	"net/http"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listfolders"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	"github.com/uptrace/bun"
)

// Handler shows a folder and its direct children: GET /api/folders/{id}.
// @Summary  Read a folder and its direct children
// @Tags     folders
// @Produce  json
// @Param    id path string true "uuid"
// @Success  200 {object} readitem.Result
// @Failure  404 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/folders/{id} [get]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := respond.ID(w, r)
		if !ok {
			return
		}

		process := readitem.Process(r.Context(), db, domain.FOLDER, id)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		folder := process.GetResult()
		folder.Children = append([]domain.Item{}, folder.Children...)
		respond.JSON(w, http.StatusOK, folder)
	}
}

// List gives every folder with its path: GET /api/folders.
// @Summary  List every folder with its path
// @Tags     folders
// @Produce  json
// @Success  200 {array} listfolders.Folder
// @Failure  500 {object} respond.Failure
// @Router   /api/folders [get]
func List(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		process := listfolders.Process(r.Context(), db)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.List(w, http.StatusOK, process.GetResult())
	}
}
