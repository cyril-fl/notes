package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/uptrace/bun"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createschema"
)

const unknownID = "3f2a4b1c-0d9e-4a8b-9c7d-1e2f3a4b5c6d"

// client calls the whole API, guard included, inside a transaction rolled back at the end.
type client struct {
	t       *testing.T
	handler http.Handler
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
	return &client{t: t, handler: New(bun.IDB(tx), "")}
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
	c.do("POST", "/api/notes", fmt.Sprintf(`{"title":%q,"content":"hello","tags":["go","sql"]}`, title), http.StatusCreated, &note)
	if note.Kind != "note" || note.Title != title {
		t.Fatalf("note = %+v", note)
	}
	path := "/api/notes/" + note.ID

	var read detail
	c.do("GET", path, "", http.StatusOK, &read)
	if read.Content != "hello\n\n#go #sql" || read.Path != "/"+title || strings.Join(read.Tags, ",") != "go,sql" {
		t.Errorf("read = %+v", read)
	}
	if strings.Contains(c.raw(path), "children") {
		t.Errorf("a note has no children field: %s", c.raw(path))
	}

	// One field changed, the others kept; the tags are those of the text.
	c.do("PATCH", path, `{"content":"changed #go"}`, http.StatusOK, nil)
	c.do("GET", path, "", http.StatusOK, &read)
	if read.Content != "changed #go" || read.Item.Title != title || strings.Join(read.Tags, ",") != "go" {
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

	c.do("PATCH", path, `{"title":""}`, http.StatusOK, nil)
	c.do("GET", path, "", http.StatusOK, &read)
	if read.Item.Title != "" || read.Content != "changed" {
		t.Errorf("after title patch: %+v", read)
	}

	c.do("PATCH", path, `{"content":" "}`, http.StatusBadRequest, nil)
	c.do("PATCH", path, `{"nope":1}`, http.StatusBadRequest, nil)
	c.do("PATCH", "/api/notes/"+unknownID, `{"content":"x"}`, http.StatusNotFound, nil)
	c.do("POST", "/api/notes", `{"title":"no content"}`, http.StatusBadRequest, nil)
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
	c.do("POST", "/api/notes", `{"title":"loose","content":"C"}`, http.StatusCreated, &note)
	c.do("POST", "/api/folders", `{"title":" "}`, http.StatusBadRequest, nil)

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
	c.do("PATCH", path, `{}`, http.StatusBadRequest, nil)
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
	c.do("POST", "/api/notes", fmt.Sprintf(`{"title":"a","content":"needle-%s","tags":[%q]}`, first, first), http.StatusCreated, &a)
	c.do("POST", "/api/notes", fmt.Sprintf(`{"title":"b","content":"C","tags":[%q]}`, first), http.StatusCreated, &b)

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
func TestRouterGuards(t *testing.T) {
	handler := New(nil, "http://localhost:3001")

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

	r = httptest.NewRequest(http.MethodPut, "/api/notes/"+unknownID, nil)
	r.Host = "localhost:8080"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("unknown method: status = %d, want 405", w.Code)
	}
}
