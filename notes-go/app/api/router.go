// Package api is the HTTP API: one folder per route, each calling the same processes as the CLI.
package api

import (
	"net/http"

	apicreatefile "github.com/cyril-fl/notes/notes-go/app/api/create-file"
	apicreatefolder "github.com/cyril-fl/notes/notes-go/app/api/create-folder"
	apicreatenote "github.com/cyril-fl/notes/notes-go/app/api/create-note"
	apicreatetag "github.com/cyril-fl/notes/notes-go/app/api/create-tag"
	apideletefolder "github.com/cyril-fl/notes/notes-go/app/api/delete-folder"
	apideletenote "github.com/cyril-fl/notes/notes-go/app/api/delete-note"
	apideletetag "github.com/cyril-fl/notes/notes-go/app/api/delete-tag"
	apideletetrash "github.com/cyril-fl/notes/notes-go/app/api/delete-trash"
	apieditfolder "github.com/cyril-fl/notes/notes-go/app/api/edit-folder"
	apieditnote "github.com/cyril-fl/notes/notes-go/app/api/edit-note"
	apiedittag "github.com/cyril-fl/notes/notes-go/app/api/edit-tag"
	apiexport "github.com/cyril-fl/notes/notes-go/app/api/export"
	apiimport "github.com/cyril-fl/notes/notes-go/app/api/import"
	apilogin "github.com/cyril-fl/notes/notes-go/app/api/login"
	apimove "github.com/cyril-fl/notes/notes-go/app/api/move"
	apireadbrokenlinks "github.com/cyril-fl/notes/notes-go/app/api/read-broken-links"
	apireadfile "github.com/cyril-fl/notes/notes-go/app/api/read-file"
	apireadfiles "github.com/cyril-fl/notes/notes-go/app/api/read-files"
	apireadfinder "github.com/cyril-fl/notes/notes-go/app/api/read-finder"
	apireadfolder "github.com/cyril-fl/notes/notes-go/app/api/read-folder"
	apireadgraph "github.com/cyril-fl/notes/notes-go/app/api/read-graph"
	apireadnote "github.com/cyril-fl/notes/notes-go/app/api/read-note"
	apireadpinned "github.com/cyril-fl/notes/notes-go/app/api/read-pinned"
	apireadrecent "github.com/cyril-fl/notes/notes-go/app/api/read-recent"
	apireadsearch "github.com/cyril-fl/notes/notes-go/app/api/read-search"
	apireadtag "github.com/cyril-fl/notes/notes-go/app/api/read-tag"
	apireadtrash "github.com/cyril-fl/notes/notes-go/app/api/read-trash"
	apireadversions "github.com/cyril-fl/notes/notes-go/app/api/read-versions"
	apirestore "github.com/cyril-fl/notes/notes-go/app/api/restore"
	apisearch "github.com/cyril-fl/notes/notes-go/app/api/search"
	"github.com/cyril-fl/notes/notes-go/app/api/services/auth"
	"github.com/cyril-fl/notes/notes-go/app/api/services/guard"
	apiusers "github.com/cyril-fl/notes/notes-go/app/api/users"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/silo"
	"github.com/uptrace/bun"
)

// @title    Notes API
// @version  1.0

// New gives the API, reachable by web pages from origin only (none when empty), and by logged-in accounts only.
// The images of the notes are in store. addUserMode is how accounts are created, one of domain.AddUserModes.
func New(db bun.IDB, store silo.Store, origin, addUserMode string) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /api/login", apilogin.Handler(db))
	mux.Handle("POST /api/logout", apilogin.Logout(db))
	mux.Handle("GET /api/me", apilogin.Me(db))
	mux.Handle("PATCH /api/me", apilogin.EditMe(db))
	mux.Handle("DELETE /api/me/data", apilogin.DeleteData(db))
	mux.Handle("PATCH /api/me/password", apilogin.EditPassword(db))

	mux.Handle("GET /api/add-user", apiusers.Modes(db, addUserMode))
	mux.Handle("POST /api/signup", apiusers.Signup(db, addUserMode))
	mux.Handle("GET /api/users", apiusers.List(db))
	mux.Handle("DELETE /api/users/{id}", apiusers.Delete(db, store))
	mux.Handle("GET /api/invites", apiusers.Invites(db))
	mux.Handle("POST /api/invites", apiusers.Invite(db))
	mux.Handle("DELETE /api/invites/{id}", apiusers.Revoke(db))
	mux.Handle("GET /api/version", apiusers.VersionHandler(db))

	mux.Handle("POST /api/notes", apicreatenote.Handler(db))
	mux.Handle("GET /api/notes/{id}", apireadnote.Handler(db))
	mux.Handle("PATCH /api/notes/{id}", apieditnote.Handler(db))
	mux.Handle("DELETE /api/notes/{id}", apideletenote.Handler(db))
	mux.Handle("GET /api/notes/{id}/neighbors", apireadgraph.Neighbors(db))
	mux.Handle("GET /api/notes/{id}/versions", apireadversions.Handler(db))
	mux.Handle("POST /api/notes/{id}/versions", apireadversions.Keep(db))
	mux.Handle("GET /api/notes/{id}/versions/{version}", apireadversions.Read(db))

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
	mux.Handle("GET /api/links/broken", apireadbrokenlinks.Handler(db))
	mux.Handle("GET /api/search", apisearch.Handler(db))
	mux.Handle("GET /api/searches/{id}", apireadsearch.Handler(db))
	mux.Handle("GET /api/recent", apireadrecent.Handler(db))
	mux.Handle("GET /api/pinned", apireadpinned.Handler(db))
	mux.Handle("GET /api/trash", apireadtrash.Handler(db))
	mux.Handle("DELETE /api/trash", apideletetrash.Handler(db))
	mux.Handle("DELETE /api/trash/{id}", apideletetrash.Item(db))
	mux.Handle("POST /api/restore", apirestore.Handler(db))
	mux.Handle("GET /api/export", apiexport.Handler(db, store))
	mux.Handle("POST /api/import", apiimport.Handler(db, store))

	mux.Handle("POST /api/files", apicreatefile.Handler(db, store))
	mux.Handle("GET /api/files", apireadfiles.Handler(db))
	mux.Handle("GET /api/files/{id}", apireadfile.Handler(db, store))

	return guard.New(origin, auth.New(db, mux))
}
