package readfolder

import (
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/process/listfolders"
	"main.go/app/shared/services/process/readitem"
)

// Handler shows a folder and its direct children: GET /api/folders/{id}.
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
