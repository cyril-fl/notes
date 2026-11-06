package editfolder

import (
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/editfolder"
)

// Handler renames a folder: PATCH /api/folders/{id}.
// @Summary  Rename a folder
// @Tags     folders
// @Accept   json
// @Produce  json
// @Param    id path string true "uuid"
// @Param    body body editfolder.Input true "body"
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

		var input editfolder.Input
		if !respond.Decode(w, r, &input) {
			return
		}
		input.ID = id

		process := editfolder.Process(r.Context(), db, input)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.JSON(w, http.StatusOK, process.GetResult())
	}
}
