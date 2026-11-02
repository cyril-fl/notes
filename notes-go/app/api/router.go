// Package api is the HTTP API: one folder per route, each calling the same processes as the CLI.
package api

import (
	"net/http"

	"github.com/uptrace/bun"
	apicreatenote "main.go/app/api/create-note"
	apideletenote "main.go/app/api/delete-note"
	apieditnote "main.go/app/api/edit-note"
	apireadnote "main.go/app/api/read-note"
	"main.go/app/api/services/guard"
)

// New gives the API, reachable by web pages from origin only (none when empty).
func New(db bun.IDB, origin string) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /api/notes", apicreatenote.Handler(db))
	mux.Handle("GET /api/notes/{id}", apireadnote.Handler(db))
	mux.Handle("PATCH /api/notes/{id}", apieditnote.Handler(db))
	mux.Handle("DELETE /api/notes/{id}", apideletenote.Handler(db))

	return guard.New(origin, mux)
}
