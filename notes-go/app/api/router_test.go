package api

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/password"
	"main.go/app/shared/services/process/adduser"
	"main.go/app/shared/services/testaccount"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/uptrace/bun"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/silo"
)

const unknownID = "3f2a4b1c-0d9e-4a8b-9c7d-1e2f3a4b5c6d"

// client calls the whole API, guard included, inside a transaction rolled back at the end.
type client struct {
	t       *testing.T
	handler http.Handler // logged in
	api     http.Handler // as a request without session gets it
	db      bun.Tx
}

func newClient(t *testing.T) *client {
	t.Helper()
	ctx := context.Background()
	env := postgrey.NewEnvironment()
	env.ReadEnv()
	db := env.Connect()
	t.Cleanup(func() { db.Close() })

	schema := createschema.Process(ctx, db)
	schema.Run()
	if schema.HasErr() {
		t.Fatalf("create schema: %v", schema.GetErr())
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { tx.Rollback() })
	return logIn(t, tx)
}

// other is a client of another account, in the same base.
func (c *client) other(t *testing.T) *client {
	return logIn(t, c.db)
}

// logIn makes an account in tx and a client logged in as it: a session made straight in the base.
func logIn(t *testing.T, tx bun.Tx) *client {
	t.Helper()
	ctx, savepoint := testaccount.New(t, context.Background(), tx)
	token := password.Token()
	session := domain.Session{UserID: domain.UserID(ctx), TokenHash: password.TokenHash(token), ExpiresAt: time.Now().Add(time.Hour)}
	if _, err := savepoint.NewInsert().Model(&session).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	api := New(bun.IDB(savepoint), silo.NewMemory(), "")
	return &client{t: t, api: api, db: savepoint, handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Authorization", "Bearer "+token)
		api.ServeHTTP(w, r)
	})}
}

// do sends the request and checks the status; dest receives the JSON answer when given.
func (c *client) do(method, path, body string, status int, dest any) {
	c.t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Host = "localhost:8080"

	w := httptest.NewRecorder()
	c.handler.ServeHTTP(w, r)
	if w.Code != status {
		c.t.Fatalf("%s %s: status = %d, want %d: %s", method, path, w.Code, status, w.Body)
	}
	if dest != nil {
		if err := json.Unmarshal(w.Body.Bytes(), dest); err != nil {
			c.t.Fatalf("%s %s: %v: %s", method, path, err, w.Body)
		}
	}
}

// raw returns the answer as text, to check how empty lists are written.
func (c *client) raw(path string) string {
	c.t.Helper()
	r := httptest.NewRequest(http.MethodGet, path, nil)
	r.Host = "localhost:8080"
	w := httptest.NewRecorder()
	c.handler.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		c.t.Fatalf("GET %s: status = %d: %s", path, w.Code, w.Body)
	}
	return w.Body.String()
}

type item struct {
	ID        string    `json:"id"`
	ParentID  *string   `json:"parent_id"`
	CreatedAt time.Time `json:"created_at"`
	Title     string    `json:"title"`
	Kind      string    `json:"kind"`
}

type detail struct {
	Item     item     `json:"item"`
	Path     string   `json:"path"`
	Content  string   `json:"content"`
	Tags     []string `json:"tags"`
	Children []item   `json:"children"`
}

func unique(name string) string {
	return fmt.Sprintf("%s-%d", name, time.Now().UnixNano())
}

func TestNotes(t *testing.T) {
	c := newClient(t)
	title := unique("note")

	var note item
	// The title is the first heading of the content.
	c.do("POST", "/api/notes", fmt.Sprintf(`{"content":"# %s\n\nhello","tags":["go","sql"]}`, title), http.StatusCreated, &note)
	if note.Kind != "note" || note.Title != title {
		t.Fatalf("note = %+v", note)
	}
	path := "/api/notes/" + note.ID

	var read detail
	c.do("GET", path, "", http.StatusOK, &read)
	if read.Content != "# "+title+"\n\nhello\n\n#go #sql" || read.Path != "/"+title || strings.Join(read.Tags, ",") != "go,sql" {
		t.Errorf("read = %+v", read)
	}
	if strings.Contains(c.raw(path), "children") {
		t.Errorf("a note has no children field: %s", c.raw(path))
	}

	// One field changed, the others kept; the tags are those of the text, the title its first words.
	c.do("PATCH", path, `{"content":"changed #go"}`, http.StatusOK, nil)
	c.do("GET", path, "", http.StatusOK, &read)
	if read.Content != "changed #go" || read.Item.Title != "changed" || strings.Join(read.Tags, ",") != "go" {
		t.Errorf("after content patch: %+v", read)
	}

	c.do("PATCH", path, `{"tags":["go","sql"]}`, http.StatusOK, nil)
	c.do("GET", path, "", http.StatusOK, &read)
	if read.Content != "changed #go\n\n#sql" || strings.Join(read.Tags, ",") != "go,sql" {
		t.Errorf("tags are written into the text: %+v", read)
	}

	c.do("PATCH", path, `{"tags":["a b"]}`, http.StatusBadRequest, nil)

	c.do("PATCH", path, `{"tags":[]}`, http.StatusOK, nil)
	if !strings.Contains(c.raw(path), `"tags":[]`) || strings.Contains(c.raw(path), "#") {
		t.Errorf("an explicit empty list clears the tags and their text: %s", c.raw(path))
	}

	// The title is not given: it comes from the content.
	c.do("PATCH", path, `{"title":"x"}`, http.StatusBadRequest, nil)
	c.do("GET", path, "", http.StatusOK, &read)
	if read.Item.Title != "changed" || read.Content != "changed" {
		t.Errorf("after tags cleared: %+v", read)
	}

	// Pinning alone leaves the text as it is.
	c.do("PATCH", path, `{"pinned":true}`, http.StatusOK, nil)
	c.do("GET", path, "", http.StatusOK, &read)
	if !strings.Contains(c.raw(path), `"pinned":true`) || read.Content != "changed" || !strings.Contains(c.raw("/api/pinned"), note.ID) {
		t.Errorf("after pin: %s", c.raw(path))
	}
	c.do("PATCH", path, `{"pinned":false}`, http.StatusOK, nil)
	if strings.Contains(c.raw("/api/pinned"), note.ID) {
		t.Errorf("an unpinned note is listed: %s", c.raw("/api/pinned"))
	}

	c.do("PATCH", path, `{"content":" "}`, http.StatusBadRequest, nil)
	c.do("PATCH", path, `{"nope":1}`, http.StatusBadRequest, nil)
	c.do("PATCH", "/api/notes/"+unknownID, `{"content":"x"}`, http.StatusNotFound, nil)
	c.do("POST", "/api/notes", `{"content":""}`, http.StatusBadRequest, nil)
	c.do("POST", "/api/notes", `{"content":"x","parent_id":"nope"}`, http.StatusBadRequest, nil)
	c.do("POST", "/api/notes", fmt.Sprintf(`{"content":"x","parent_id":%q}`, unknownID), http.StatusBadRequest, nil)
	c.do("GET", "/api/notes/"+unknownID, "", http.StatusNotFound, nil)
	c.do("GET", "/api/notes/nope", "", http.StatusNotFound, nil)

	c.do("DELETE", path, "", http.StatusNoContent, nil)
	c.do("GET", path, "", http.StatusNotFound, nil)
	c.do("DELETE", path, "", http.StatusNotFound, nil)
	c.do("DELETE", "/api/notes/nope", "", http.StatusNotFound, nil)

	var restored []item
	c.do("POST", "/api/restore", fmt.Sprintf(`{"ids":[%q]}`, note.ID), http.StatusOK, &restored)
	if len(restored) != 1 || restored[0].ID != note.ID || restored[0].CreatedAt.IsZero() {
		t.Errorf("restored = %+v", restored)
	}
	c.do("GET", path, "", http.StatusOK, nil)
	c.do("POST", "/api/restore", `{"ids":[]}`, http.StatusBadRequest, nil)
	c.do("POST", "/api/restore", `{"ids":["nope"]}`, http.StatusBadRequest, nil)
}

func TestFoldersAndMove(t *testing.T) {
	c := newClient(t)
	title := unique("folder")

	var folder, sub, note item
	c.do("POST", "/api/folders", fmt.Sprintf(`{"title":%q}`, title), http.StatusCreated, &folder)
	path := "/api/folders/" + folder.ID

	if !strings.Contains(c.raw(path), `"children":[]`) {
		t.Errorf("an empty folder has children: []: %s", c.raw(path))
	}

	c.do("POST", "/api/folders", fmt.Sprintf(`{"title":"sub","parent_id":%q}`, folder.ID), http.StatusCreated, &sub)
	c.do("POST", "/api/notes", `{"content":"# loose"}`, http.StatusCreated, &note)
	c.do("POST", "/api/folders", `{"title":" "}`, http.StatusBadRequest, nil)
	c.do("POST", "/api/folders", fmt.Sprintf(`{"title":"SUB","parent_id":%q}`, folder.ID), http.StatusConflict, nil)

	// A favorite keeps its title, and a note cannot be one through the folders.
	c.do("PATCH", path, `{"pinned":true}`, http.StatusOK, nil)
	if !strings.Contains(c.raw("/api/pinned"), folder.ID) || !strings.Contains(c.raw(path), title) {
		t.Errorf("the favorite folder is not listed, or lost its title: %s", c.raw("/api/pinned"))
	}
	c.do("PATCH", "/api/folders/"+note.ID, `{"pinned":true}`, http.StatusNotFound, nil)

	// A note and a folder are not interchangeable.
	c.do("GET", "/api/notes/"+folder.ID, "", http.StatusNotFound, nil)
	c.do("GET", "/api/folders/"+note.ID, "", http.StatusNotFound, nil)
	c.do("DELETE", "/api/notes/"+folder.ID, "", http.StatusNotFound, nil)
	c.do("DELETE", "/api/folders/"+note.ID, "", http.StatusNotFound, nil)

	var moved []item
	c.do("POST", "/api/move", fmt.Sprintf(`{"ids":[%q],"parent_id":%q}`, note.ID, folder.ID), http.StatusOK, &moved)
	if len(moved) != 1 || moved[0].ParentID == nil || *moved[0].ParentID != folder.ID || moved[0].CreatedAt.IsZero() {
		t.Errorf("moved = %+v", moved)
	}

	var read detail
	c.do("GET", path, "", http.StatusOK, &read)
	if read.Path != "/"+title || len(read.Children) != 2 {
		t.Errorf("read = %+v", read)
	}

	var items []item
	c.do("GET", "/api/items?parent="+folder.ID, "", http.StatusOK, &items)
	if len(items) != 2 || items[0].Kind != "folder" {
		t.Errorf("items = %+v", items)
	}
	c.do("GET", "/api/items", "", http.StatusOK, &items)
	found := false
	for _, root := range items {
		found = found || root.ID == folder.ID
	}
	if !found {
		t.Errorf("the root does not list the folder: %+v", items)
	}
	if got := c.raw("/api/items?parent=" + note.ID); strings.TrimSpace(got) != "[]" {
		t.Errorf("no children = %q, want []", got)
	}
	c.do("GET", "/api/items?parent=nope", "", http.StatusBadRequest, nil)

	var folders []struct {
		ID   string `json:"id"`
		Path string `json:"path"`
	}
	c.do("GET", "/api/folders", "", http.StatusOK, &folders)
	found = false
	for _, f := range folders {
		found = found || f.ID == sub.ID && f.Path == "/"+title+"/sub"
	}
	if !found {
		t.Errorf("folders = %+v", folders)
	}

	renamed := unique("renamed")
	c.do("PATCH", path, fmt.Sprintf(`{"title":%q}`, renamed), http.StatusOK, &folder)
	if folder.Title != renamed {
		t.Errorf("folder = %+v", folder)
	}
	c.do("PATCH", path, `{"title":""}`, http.StatusBadRequest, nil)
	c.do("PATCH", "/api/folders/"+unknownID, `{"title":"x"}`, http.StatusNotFound, nil)

	c.do("POST", "/api/move", fmt.Sprintf(`{"ids":[%q],"parent_id":%q}`, folder.ID, sub.ID), http.StatusConflict, nil)
	c.do("POST", "/api/move", fmt.Sprintf(`{"ids":[%q],"parent_id":%q}`, note.ID, unknownID), http.StatusBadRequest, nil)
	c.do("POST", "/api/move", `{"ids":[]}`, http.StatusBadRequest, nil)
	c.do("POST", "/api/move", `{"ids":["nope"]}`, http.StatusBadRequest, nil)

	// Removing a folder takes its content, and the trash shows it once.
	c.do("DELETE", path, "", http.StatusNoContent, nil)
	c.do("GET", "/api/notes/"+note.ID, "", http.StatusNotFound, nil)
	var trash []struct {
		ID     string `json:"id"`
		Inside int    `json:"inside"`
	}
	c.do("GET", "/api/trash", "", http.StatusOK, &trash)
	found = false
	for _, removed := range trash {
		found = found || removed.ID == folder.ID && removed.Inside == 2
	}
	if !found {
		t.Errorf("trash = %+v", trash)
	}

	// A child cannot come back while its folder is in the trash.
	var parentless []item
	c.do("POST", "/api/restore", fmt.Sprintf(`{"ids":[%q]}`, note.ID), http.StatusConflict, nil)
	c.do("POST", "/api/restore", fmt.Sprintf(`{"ids":[%q]}`, folder.ID), http.StatusOK, &parentless)
	if len(parentless) != 3 {
		t.Errorf("restored = %+v", parentless)
	}
}

func TestEmptyTrash(t *testing.T) {
	c := newClient(t)

	var folder, inside, removed, live item
	c.do("POST", "/api/folders", fmt.Sprintf(`{"title":%q}`, unique("folder")), http.StatusCreated, &folder)
	c.do("POST", "/api/notes", fmt.Sprintf(`{"content":"C","parent_id":%q}`, folder.ID), http.StatusCreated, &inside)
	c.do("POST", "/api/notes", `{"content":"C"}`, http.StatusCreated, &removed)
	c.do("POST", "/api/notes", `{"content":"C"}`, http.StatusCreated, &live)
	c.do("DELETE", "/api/folders/"+folder.ID, "", http.StatusNoContent, nil)
	c.do("DELETE", "/api/notes/"+removed.ID, "", http.StatusNoContent, nil)

	// One removal goes alone, its content with it; a live item is not in the trash.
	c.do("DELETE", "/api/trash/"+folder.ID, "", http.StatusNoContent, nil)
	if got := c.raw("/api/trash"); strings.Contains(got, folder.ID) || !strings.Contains(got, removed.ID) {
		t.Errorf("trash = %s, want the note only", got)
	}
	c.do("DELETE", "/api/trash/"+live.ID, "", http.StatusNotFound, nil)
	c.do("DELETE", "/api/trash/nope", "", http.StatusNotFound, nil)

	c.do("DELETE", "/api/trash", "", http.StatusNoContent, nil)

	if got := c.raw("/api/trash"); strings.TrimSpace(got) != "[]" {
		t.Errorf("trash = %q, want []", got)
	}
	c.do("GET", "/api/notes/"+live.ID, "", http.StatusOK, nil)
	// Gone for good: nothing left to restore, the folder's content included.
	var restored []item
	c.do("POST", "/api/restore", fmt.Sprintf(`{"ids":[%q,%q,%q]}`, folder.ID, inside.ID, removed.ID), http.StatusOK, &restored)
	if len(restored) != 0 {
		t.Errorf("restored = %+v", restored)
	}

	c.do("DELETE", "/api/trash", "", http.StatusNoContent, nil)
}

// The data of the account go, the account and its session stay; another account keeps its own.
func TestDeleteData(t *testing.T) {
	alice := newClient(t)
	var kept item
	alice.do("POST", "/api/notes", `{"content":"# Kept"}`, http.StatusCreated, &kept)

	bob := alice.other(t)
	var gone item
	bob.do("POST", "/api/notes", `{"content":"# Gone #bob"}`, http.StatusCreated, &gone)
	bob.do("DELETE", "/api/me/data", "", http.StatusNoContent, nil)

	bob.do("GET", "/api/notes/"+gone.ID, "", http.StatusNotFound, nil)
	if got := bob.raw("/api/tags"); strings.TrimSpace(got) != "[]" {
		t.Errorf("tags = %q, want []", got)
	}
	bob.do("GET", "/api/me", "", http.StatusOK, nil)
	alice.do("GET", "/api/notes/"+kept.ID, "", http.StatusOK, nil)
}

func TestTagsAndGraph(t *testing.T) {
	c := newClient(t)
	first, second := unique("first"), unique("second")

	var created []struct {
		ID    string `json:"id"`
		Value string `json:"value"`
	}
	c.do("POST", "/api/tags", fmt.Sprintf(`{"values":[%q,%q]}`, first, second), http.StatusCreated, &created)
	if len(created) != 2 {
		t.Fatalf("created = %+v", created)
	}
	// Already there: skipped, an empty list and not null.
	if body := c.post("/api/tags", fmt.Sprintf(`{"values":[%q]}`, first)); strings.TrimSpace(body) != "[]" {
		t.Errorf("duplicate = %q, want []", body)
	}
	c.do("POST", "/api/tags", `{"values":[]}`, http.StatusBadRequest, nil)

	var a, b item
	c.do("POST", "/api/notes", fmt.Sprintf(`{"content":"needle-%s","tags":[%q]}`, first, first), http.StatusCreated, &a)
	c.do("POST", "/api/notes", fmt.Sprintf(`{"content":"C","tags":[%q]}`, first), http.StatusCreated, &b)

	var tag struct {
		Tag struct {
			Value string `json:"value"`
		} `json:"tag"`
		Notes []struct {
			ID   string `json:"id"`
			Path string `json:"path"`
		} `json:"notes"`
	}
	c.do("GET", "/api/tags/"+first, "", http.StatusOK, &tag)
	if tag.Tag.Value != first || len(tag.Notes) != 2 {
		t.Errorf("tag = %+v", tag)
	}
	if !strings.Contains(c.raw("/api/tags/"+second), `"notes":[]`) {
		t.Errorf("an unused tag has notes: []: %s", c.raw("/api/tags/"+second))
	}
	c.do("GET", "/api/tags/"+unique("unknown"), "", http.StatusNotFound, nil)

	var neighbors struct {
		Note      detail `json:"note"`
		Neighbors []struct {
			ID     string   `json:"id"`
			Tags   []string `json:"tags"`
			Shared int      `json:"shared"`
		} `json:"neighbors"`
	}
	c.do("GET", "/api/notes/"+a.ID+"/neighbors", "", http.StatusOK, &neighbors)
	if neighbors.Note.Item.ID != a.ID || len(neighbors.Neighbors) != 1 || neighbors.Neighbors[0].ID != b.ID || neighbors.Neighbors[0].Shared != 1 {
		t.Errorf("neighbors = %+v", neighbors)
	}
	if raw := c.raw("/api/notes/" + a.ID + "/neighbors"); !strings.Contains(raw, `"links":[]`) || !strings.Contains(raw, `"backlinks":[]`) {
		t.Errorf("a note without [[links]] has links: [] and backlinks: []: %s", raw)
	}
	c.do("GET", "/api/notes/"+unknownID+"/neighbors", "", http.StatusNotFound, nil)

	var graph []struct {
		Value string `json:"value"`
		Notes int    `json:"notes"`
	}
	c.do("GET", "/api/graph", "", http.StatusOK, &graph)
	found := false
	for _, count := range graph {
		found = found || count.Value == first && count.Notes == 2
	}
	if !found {
		t.Errorf("graph has no %s with 2 notes: %+v", first, graph)
	}

	var notes struct {
		Notes []struct {
			ID string `json:"id"`
		} `json:"notes"`
		Links []struct {
			Source string   `json:"source"`
			Target string   `json:"target"`
			Tags   []string `json:"tags"`
		} `json:"links"`
	}
	c.do("GET", "/api/graph/notes", "", http.StatusOK, &notes)
	linked := false
	for _, link := range notes.Links {
		linked = linked || (link.Source == a.ID && link.Target == b.ID || link.Source == b.ID && link.Target == a.ID) && slices.Equal(link.Tags, []string{first})
	}
	if !linked {
		t.Errorf("graph notes have no link between a and b on %s: %+v", first, notes.Links)
	}

	var all []struct {
		Value string `json:"value"`
	}
	c.do("GET", "/api/tags", "", http.StatusOK, &all)
	found = false
	for _, listed := range all {
		found = found || listed.Value == second
	}
	if !found {
		t.Errorf("tags = %+v", all)
	}

	var matches []struct {
		ID      string `json:"id"`
		Snippet string `json:"snippet"`
	}
	c.do("GET", "/api/search?q=needle-"+first, "", http.StatusOK, &matches)
	if len(matches) != 1 || matches[0].ID != a.ID {
		t.Errorf("matches = %+v", matches)
	}
	if got := c.raw("/api/search?q=" + unique("nothing")); strings.TrimSpace(got) != "[]" {
		t.Errorf("no match = %q, want []", got)
	}
	c.do("GET", "/api/search", "", http.StatusBadRequest, nil)

	var recent []struct {
		ID string `json:"id"`
	}
	c.do("GET", "/api/recent", "", http.StatusOK, &recent)
	if len(recent) == 0 {
		t.Error("recent is empty")
	}

	renamed := unique("renamed")
	var edited struct {
		Value string `json:"value"`
	}
	c.do("PATCH", "/api/tags/"+second, fmt.Sprintf(`{"value":%q}`, renamed), http.StatusOK, &edited)
	if edited.Value != renamed {
		t.Errorf("edited = %+v", edited)
	}
	c.do("PATCH", "/api/tags/"+unique("unknown"), `{"value":"x"}`, http.StatusNotFound, nil)

	c.do("DELETE", "/api/tags/"+renamed, "", http.StatusNoContent, nil)
	c.do("DELETE", "/api/tags/"+renamed, "", http.StatusNotFound, nil)

	// Last: the unique violation aborts the transaction.
	c.do("PATCH", "/api/tags/"+first, fmt.Sprintf(`{"value":%q}`, first+"-taken"), http.StatusOK, nil)
	c.do("POST", "/api/tags", fmt.Sprintf(`{"values":[%q]}`, first), http.StatusCreated, nil)
	c.do("PATCH", "/api/tags/"+first, fmt.Sprintf(`{"value":%q}`, first+"-taken"), http.StatusConflict, nil)
}

// post returns the answer of a creation as text.
func (c *client) post(path, body string) string {
	c.t.Helper()
	r := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	r.Host = "localhost:8080"
	w := httptest.NewRecorder()
	c.handler.ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		c.t.Fatalf("POST %s: status = %d: %s", path, w.Code, w.Body)
	}
	return w.Body.String()
}

// No DB needed: the guard answers first, and the route of an unknown method never runs.
func TestExportAndImport(t *testing.T) {
	c := newClient(t)
	folder := fmt.Sprintf("transfer-%d", time.Now().UnixNano())

	var zipped bytes.Buffer
	archive := zip.NewWriter(&zipped)
	for name, content := range map[string]string{folder + "/Idées.md": "# Idées " + folder, folder + "/image.png": "png"} {
		entry, _ := archive.Create(name)
		entry.Write([]byte(content))
	}
	archive.Close()

	var result struct{ Folders, Notes int }
	c.do(http.MethodPost, "/api/import", zipped.String(), http.StatusOK, &result)
	if result.Folders != 1 || result.Notes != 1 {
		t.Errorf("import = %+v, want 1 folder and 1 note", result)
	}
	// The same folder again at the root: refused, all or nothing.
	c.do(http.MethodPost, "/api/import", zipped.String(), http.StatusConflict, nil)
	c.do(http.MethodPost, "/api/import", "pas un zip", http.StatusBadRequest, nil)

	// Into a folder, the same zip goes in again; a parent that is not a folder is refused.
	var into item
	c.do(http.MethodPost, "/api/folders", fmt.Sprintf(`{"title":%q}`, folder+"-into"), http.StatusCreated, &into)
	c.do(http.MethodPost, "/api/import?parent="+into.ID, zipped.String(), http.StatusOK, &result)
	c.do(http.MethodPost, "/api/import?parent="+unknownID, zipped.String(), http.StatusBadRequest, nil)
	c.do(http.MethodPost, "/api/import?parent=nope", zipped.String(), http.StatusBadRequest, nil)

	exported := c.raw("/api/export")
	read, err := zip.NewReader(strings.NewReader(exported), int64(len(exported)))
	if err != nil {
		t.Fatalf("export is not a zip: %v", err)
	}
	if file, err := read.Open(folder + "/Idées " + folder + ".md"); err != nil {
		t.Errorf("export misses the imported note: %v", err)
	} else {
		file.Close()
	}
}

func TestRouterGuards(t *testing.T) {
	handler := New(nil, nil, "http://localhost:3001")

	r := httptest.NewRequest(http.MethodDelete, "/api/notes/"+unknownID, nil)
	r.Host = "localhost:8080"
	r.Header.Set("Origin", "https://evil.example")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Errorf("other origin: status = %d, want 403", w.Code)
	}

	r = httptest.NewRequest(http.MethodGet, "/api/notes/"+unknownID, nil)
	r.Host = "evil.example"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Errorf("other host: status = %d, want 403", w.Code)
	}

	// Without a session, nothing answers, not even a 405: the routes stay unknown.
	r = httptest.NewRequest(http.MethodPut, "/api/notes/"+unknownID, nil)
	r.Host = "localhost:8080"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("no session: status = %d, want 401", w.Code)
	}
}

func TestFiles(t *testing.T) {
	c := newClient(t)
	image := "\x89PNG\r\n\x1a\n...."

	var file struct{ ID, Name, ContentType string }
	c.do(http.MethodPost, "/api/files?name=chat.png", image, http.StatusCreated, &file)
	if file.ID == "" || file.Name != "chat.png" {
		t.Fatalf("file = %+v", file)
	}
	if got := c.raw("/api/files/" + file.ID); got != image {
		t.Errorf("GET = %q, want the image back", got)
	}
	c.do(http.MethodPost, "/api/files", "<svg/>", http.StatusUnsupportedMediaType, nil)
	c.do(http.MethodPost, "/api/files", "", http.StatusBadRequest, nil)
	c.do(http.MethodGet, "/api/files/"+unknownID, "", http.StatusNotFound, nil)
	c.other(t).do(http.MethodGet, "/api/files/"+file.ID, "", http.StatusNotFound, nil)

	// A note citing it exports the image beside it.
	title := unique("photo")
	c.do(http.MethodPost, "/api/notes", fmt.Sprintf(`{"content":"# %s\n\n![chat](/api/files/%s)"}`, title, file.ID), http.StatusCreated, nil)
	exported := c.raw("/api/export")
	read, err := zip.NewReader(strings.NewReader(exported), int64(len(exported)))
	if err != nil {
		t.Fatalf("export is not a zip: %v", err)
	}
	for name, want := range map[string]string{title + ".md": "![chat](_files/" + file.ID + ".png)", "_files/" + file.ID + ".png": image} {
		entry, err := read.Open(name)
		if err != nil {
			t.Fatalf("export misses %s: %v", name, err)
		}
		data, _ := io.ReadAll(entry)
		entry.Close()
		if !strings.Contains(string(data), want) {
			t.Errorf("%s = %q, want %q in it", name, data, want)
		}
	}
}

// Two accounts in one base: neither sees, changes nor links to what the other has.
func TestAccountsAreApart(t *testing.T) {
	alice := newClient(t)
	var folder, note item
	alice.do("POST", "/api/folders", fmt.Sprintf(`{"title":%q}`, unique("alice")), http.StatusCreated, &folder)
	alice.do("POST", "/api/notes", fmt.Sprintf(`{"content":"# Secret #privé","parent_id":%q}`, folder.ID), http.StatusCreated, &note)

	bob := alice.other(t)
	bob.do("GET", "/api/notes/"+note.ID, "", http.StatusNotFound, nil)
	bob.do("GET", "/api/folders/"+folder.ID, "", http.StatusNotFound, nil)
	bob.do("PATCH", "/api/notes/"+note.ID, `{"content":"# volé"}`, http.StatusNotFound, nil)
	bob.do("DELETE", "/api/notes/"+note.ID, "", http.StatusNotFound, nil)
	bob.do("POST", "/api/notes", fmt.Sprintf(`{"content":"# Intrus","parent_id":%q}`, folder.ID), http.StatusBadRequest, nil)
	for _, path := range []string{"/api/items", "/api/search?q=Secret", "/api/graph", "/api/graph/notes", "/api/recent", "/api/folders", "/api/tags"} {
		if got := bob.raw(path); strings.Contains(got, note.ID) || strings.Contains(got, folder.ID) || strings.Contains(got, "privé") {
			t.Errorf("%s shows alice's data to bob: %s", path, got)
		}
	}
	// A link by id to a note of another account links nothing.
	var linking item
	bob.do("POST", "/api/notes", fmt.Sprintf(`{"content":"# Lien\n\n[x](note:%s)"}`, note.ID), http.StatusCreated, &linking)
	if got := bob.raw("/api/notes/" + linking.ID + "/neighbors"); !strings.Contains(got, `"links":[]`) {
		t.Errorf("bob links to alice's note: %s", got)
	}
}

func TestLogin(t *testing.T) {
	c := newClient(t)
	name := unique("login")
	add := adduser.Process(context.Background(), c.db, adduser.Input{Name: name, Password: "correct horse"})
	add.Run()
	if add.HasErr() {
		t.Fatal(add.GetErr())
	}

	send := func(method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Host = "localhost:8080"
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		c.api.ServeHTTP(w, r)
		return w
	}

	if w := send("POST", "/api/login", fmt.Sprintf(`{"name":%q,"password":"wrong horse"}`, name), nil); w.Code != http.StatusUnauthorized {
		t.Errorf("wrong password: status = %d, want 401", w.Code)
	}
	w := send("POST", "/api/login", fmt.Sprintf(`{"name":%q,"password":"correct horse"}`, name), nil)
	cookies := w.Result().Cookies()
	if w.Code != http.StatusOK || len(cookies) != 1 || !cookies[0].HttpOnly || strings.Contains(w.Body.String(), cookies[0].Value) {
		t.Fatalf("login: status = %d, cookies = %v, body = %s; want an HttpOnly cookie, the token out of the body", w.Code, cookies, w.Body)
	}
	if w := send("GET", "/api/me", "", cookies[0]); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), name) {
		t.Errorf("me: status = %d, body = %s", w.Code, w.Body)
	}
	if w := send("PATCH", "/api/me", `{"locale":"de"}`, cookies[0]); w.Code != http.StatusBadRequest {
		t.Errorf("unknown locale: status = %d, want 400", w.Code)
	}
	if w := send("PATCH", "/api/me", `{"locale":"en"}`, cookies[0]); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"locale":"en"`) {
		t.Errorf("edit me: status = %d, body = %s", w.Code, w.Body)
	}
	if w := send("PATCH", "/api/me", `{"trash_retention_days":-1}`, cookies[0]); w.Code != http.StatusBadRequest {
		t.Errorf("negative retention: status = %d, want 400", w.Code)
	}
	if w := send("PATCH", "/api/me", `{"theme":"dracula","color_mode":"dark","view":"grid","trash_retention_days":7}`, cookies[0]); w.Code != http.StatusOK ||
		!strings.Contains(w.Body.String(), `"locale":"en","theme":"dracula","color_mode":"dark","view":"grid","trash_retention_days":7`) {
		t.Errorf("edit settings: status = %d, body = %s", w.Code, w.Body)
	}
	if w := send("PATCH", "/api/me", `{"snippets":[{"name":"","content":"x"}]}`, cookies[0]); w.Code != http.StatusBadRequest {
		t.Errorf("snippet without a name: status = %d, want 400", w.Code)
	}
	if w := send("PATCH", "/api/me", `{"snippets":[{"name":"Day","content":"# Today #daily"}]}`, cookies[0]); w.Code != http.StatusOK ||
		!strings.Contains(w.Body.String(), `"snippets":[{"name":"Day","content":"# Today #daily"}]`) {
		t.Errorf("edit snippets: status = %d, body = %s", w.Code, w.Body)
	}
	if w := send("PATCH", "/api/me", `{"avatar":"00000000-0000-0000-0000-000000000000"}`, cookies[0]); w.Code != http.StatusBadRequest {
		t.Errorf("avatar of no image: status = %d, want 400", w.Code)
	}
	name += "-renamed"
	if w := send("PATCH", "/api/me", fmt.Sprintf(`{"name":%q}`, name), cookies[0]); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), name) {
		t.Errorf("rename: status = %d, body = %s", w.Code, w.Body)
	}

	// A new password logs the other devices out, not this one.
	other := send("POST", "/api/login", fmt.Sprintf(`{"name":%q,"password":"correct horse"}`, name), nil).Result().Cookies()
	if len(other) != 1 {
		t.Fatalf("second login: cookies = %v", other)
	}
	if w := send("PATCH", "/api/me/password", `{"current":"wrong horse","new":"battery staple"}`, cookies[0]); w.Code != http.StatusBadRequest {
		t.Errorf("wrong current password: status = %d, want 400", w.Code)
	}
	if w := send("PATCH", "/api/me/password", `{"current":"correct horse","new":"battery staple"}`, cookies[0]); w.Code != http.StatusNoContent {
		t.Errorf("new password: status = %d, body = %s", w.Code, w.Body)
	}
	if w := send("GET", "/api/me", "", other[0]); w.Code != http.StatusUnauthorized {
		t.Errorf("other device after a new password: status = %d, want 401", w.Code)
	}
	if w := send("GET", "/api/me", "", cookies[0]); w.Code != http.StatusOK {
		t.Errorf("this device after a new password: status = %d, want 200", w.Code)
	}
	send("POST", "/api/logout", "", cookies[0])
	if w := send("GET", "/api/me", "", cookies[0]); w.Code != http.StatusUnauthorized {
		t.Errorf("me after logout: status = %d, want 401", w.Code)
	}
}
