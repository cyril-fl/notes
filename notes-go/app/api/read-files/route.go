package readfiles

import (
	"net/http"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listfiles"
	"github.com/uptrace/bun"
)

// Handler lists the images and attachments the account keeps, the largest first: GET /api/files.
// @Summary  List the files kept on the server
// @Tags     files
// @Produce  json
// @Success  200 {array} domain.File
// @Failure  500 {object} respond.Failure
// @Router   /api/files [get]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		process := listfiles.Process(r.Context(), db)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.List(w, http.StatusOK, process.GetResult())
	}
}
