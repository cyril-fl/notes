// Package api is the HTTP API: one folder per route, each calling the same processes as the CLI.
package api

import (
	"net/http"

	"github.com/uptrace/bun"
	apicreatefolder "main.go/app/api/create-folder"
	apicreatenote "main.go/app/api/create-note"
	apideletefolder "main.go/app/api/delete-folder"
	apideletenote "main.go/app/api/delete-note"
	apieditfolder "main.go/app/api/edit-folder"
	apieditnote "main.go/app/api/edit-note"
	apimove "main.go/app/api/move"
	apireadfinder "main.go/app/api/read-finder"
	apireadfolder "main.go/app/api/read-folder"
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

	mux.Handle("GET /api/folders", apireadfolder.List(db))
	mux.Handle("POST /api/folders", apicreatefolder.Handler(db))
	mux.Handle("GET /api/folders/{id}", apireadfolder.Handler(db))
	mux.Handle("PATCH /api/folders/{id}", apieditfolder.Handler(db))
	mux.Handle("DELETE /api/folders/{id}", apideletefolder.Handler(db))

	mux.Handle("GET /api/items", apireadfinder.Handler(db))
	mux.Handle("POST /api/move", apimove.Handler(db))

	return guard.New(origin, mux)
}
