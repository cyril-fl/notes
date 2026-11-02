package readrecent

import (
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/listrecent"
)

// Handler lists the last changed notes: GET /api/recent.
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		process := listrecent.Process(r.Context(), db)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.List(w, http.StatusOK, process.GetResult())
	}
}
