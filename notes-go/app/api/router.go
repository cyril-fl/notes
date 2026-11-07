// Package api is the HTTP API: one folder per route, each calling the same processes as the CLI.
package api

import (
	"net/http"

	"github.com/uptrace/bun"
	apicreatefolder "main.go/app/api/create-folder"
	apicreatenote "main.go/app/api/create-note"
	apicreatetag "main.go/app/api/create-tag"
	apideletefolder "main.go/app/api/delete-folder"
	apideletenote "main.go/app/api/delete-note"
	apideletetag "main.go/app/api/delete-tag"
	apideletetrash "main.go/app/api/delete-trash"
	apieditfolder "main.go/app/api/edit-folder"
	apieditnote "main.go/app/api/edit-note"
	apiedittag "main.go/app/api/edit-tag"
	apiexport "main.go/app/api/export"
	apiimport "main.go/app/api/import"
	apilogin "main.go/app/api/login"
	apimove "main.go/app/api/move"
	apireadfinder "main.go/app/api/read-finder"
	apireadfolder "main.go/app/api/read-folder"
	apireadgraph "main.go/app/api/read-graph"
	apireadnote "main.go/app/api/read-note"
	apireadpinned "main.go/app/api/read-pinned"
	apireadrecent "main.go/app/api/read-recent"
	apireadtag "main.go/app/api/read-tag"
	apireadtrash "main.go/app/api/read-trash"
	apirestore "main.go/app/api/restore"
	apisearch "main.go/app/api/search"
	"main.go/app/api/services/auth"
	"main.go/app/api/services/guard"
)

// @title    Notes API
// @version  1.0

// New gives the API, reachable by web pages from origin only (none when empty), and by logged-in accounts only.
func New(db bun.IDB, origin string) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /api/login", apilogin.Handler(db))
	mux.Handle("POST /api/logout", apilogin.Logout(db))
	mux.Handle("GET /api/me", apilogin.Me(db))

	mux.Handle("POST /api/notes", apicreatenote.Handler(db))
	mux.Handle("GET /api/notes/{id}", apireadnote.Handler(db))
	mux.Handle("PATCH /api/notes/{id}", apieditnote.Handler(db))
	mux.Handle("DELETE /api/notes/{id}", apideletenote.Handler(db))
	mux.Handle("GET /api/notes/{id}/neighbors", apireadgraph.Neighbors(db))

	mux.Handle("GET /api/folders", apireadfolder.List(db))
	mux.Handle("POST /api/folders", apicreatefolder.Handler(db))
	mux.Handle("GET /api/folders/{id}", apireadfolder.Handler(db))
	mux.Handle("PATCH /api/folders/{id}", apieditfolder.Handler(db))
	mux.Handle("DELETE /api/folders/{id}", apideletefolder.Handler(db))

	mux.Handle("GET /api/items", apireadfinder.Handler(db))
	mux.Handle("POST /api/move", apimove.Handler(db))

	mux.Handle("GET /api/tags", apireadtag.List(db))
	mux.Handle("POST /api/tags", apicreatetag.Handler(db))
	mux.Handle("GET /api/tags/{value}", apireadtag.Handler(db))
	mux.Handle("PATCH /api/tags/{value}", apiedittag.Handler(db))
	mux.Handle("DELETE /api/tags/{value}", apideletetag.Handler(db))

	mux.Handle("GET /api/graph", apireadgraph.Handler(db))
	mux.Handle("GET /api/graph/notes", apireadgraph.Notes(db))
	mux.Handle("GET /api/search", apisearch.Handler(db))
	mux.Handle("GET /api/recent", apireadrecent.Handler(db))
	mux.Handle("GET /api/pinned", apireadpinned.Handler(db))
	mux.Handle("GET /api/trash", apireadtrash.Handler(db))
	mux.Handle("DELETE /api/trash", apideletetrash.Handler(db))
	mux.Handle("DELETE /api/trash/{id}", apideletetrash.Item(db))
	mux.Handle("POST /api/restore", apirestore.Handler(db))
	mux.Handle("GET /api/export", apiexport.Handler(db))
	mux.Handle("POST /api/import", apiimport.Handler(db))

	return guard.New(origin, auth.New(db, mux))
}
