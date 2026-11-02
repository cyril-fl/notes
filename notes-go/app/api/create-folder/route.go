package createfolder

import (
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/createfolder"
)

// Handler creates a folder: POST /api/folders.
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
