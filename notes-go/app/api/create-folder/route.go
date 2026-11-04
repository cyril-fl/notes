package createfolder

import (
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/createfolder"
)

// Handler creates a folder: POST /api/folders.
// @Summary  Create a folder
// @Tags     folders
// @Accept   json
// @Produce  json
// @Param    body body createfolder.Input true "body"
// @Success  201 {object} domain.Item
// @Failure  400 {object} respond.Failure
// @Failure  413 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/folders [post]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input createfolder.Input
		if !respond.Decode(w, r, &input) || !respond.ValidIDs(w, input.ParentID) {
			return
		}

		process := createfolder.Process(r.Context(), db, input)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.JSON(w, http.StatusCreated, process.GetResult())
	}
}
