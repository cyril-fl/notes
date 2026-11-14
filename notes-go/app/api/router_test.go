package api

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/password"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/adduser"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/resetpassword"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/silo"
	"github.com/uptrace/bun"
)

const unknownID = "3f2a4b1c-0d9e-4a8b-9c7d-1e2f3a4b5c6d"

// client calls the whole API, guard included, inside a transaction rolled back at the end.
type client struct {
	t       *testing.T
	ctx     context.Context // the account of handler, for the fixtures a test writes straight in the base
	handler http.Handler    // logged in
	api     http.Handler    // as a request without session gets it
	db      bun.Tx
	store   *silo.Memory
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
	store := silo.NewMemory()
	api := New(bun.IDB(savepoint), store, "", domain.AddUserModes[0])
	return &client{t: t, ctx: ctx, api: api, db: savepoint, store: store, handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	Pinned   bool     `json:"pinned"`
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
	// The PATCH answers the note as GET does.
	c.do("PATCH", path, `{"content":"changed #go"}`, http.StatusOK, &read)
	if read.Content != "changed #go" || read.Item.Title != "changed" || strings.Join(read.Tags, ",") != "go" || read.Path != "/changed" {
		t.Errorf("after content patch: %+v", read)
	}
	// The text replaced is kept as a version, the next edits of the session are not (#61, #99).
	c.do("PATCH", path, `{"content":"changed again #go"}`, http.StatusOK, nil)
	c.do("PATCH", path, `{"content":"changed #go"}`, http.StatusOK, nil)
	var versions []struct {
		ID      string `json:"id"`
		Content string `json:"content"`
	}
	c.do("GET", path+"/versions", "", http.StatusOK, &versions)
	if len(versions) != 1 || versions[0].Content != "" {
		t.Errorf("versions = %+v, want 1 without its text", versions)
	}
	c.do("GET", "/api/notes/"+unknownID+"/versions", "", http.StatusNotFound, nil)
	// The text of a version is read one at a time.
	var version struct {
		Content string `json:"content"`
	}
	c.do("GET", path+"/versions/"+versions[0].ID, "", http.StatusOK, &version)
	if version.Content != "# "+title+"\n\nhello\n\n#go #sql" {
		t.Errorf("version = %+v", version)
	}
	c.do("GET", path+"/versions/"+unknownID, "", http.StatusNotFound, nil)
	c.do("GET", path+"/versions/nope", "", http.StatusBadRequest, nil)
	// A text refused by a 409 is kept as a version marked conflict (#78); the note's own text keeps nothing.
	var kept struct {
		Content  string `json:"content"`
		Conflict bool   `json:"conflict"`
	}
	c.do("POST", path+"/versions", `{"content":"mine"}`, http.StatusCreated, &kept)
	if kept.Content != "mine" || !kept.Conflict {
		t.Errorf("conflict kept = %+v", kept)
	}
	c.do("POST", path+"/versions", `{"content":"changed #go"}`, http.StatusNoContent, nil)
	c.do("POST", "/api/notes/"+unknownID+"/versions", `{"content":"mine"}`, http.StatusNotFound, nil)

	c.do("PATCH", path, `{"tags":["go","sql"]}`, http.StatusOK, &read)
	if read.Content != "changed #go\n\n#sql" || strings.Join(read.Tags, ",") != "go,sql" {
		t.Errorf("tags are written into the text: %+v", read)
	}

	c.do("PATCH", path, `{"tags":["a b"]}`, http.StatusBadRequest, nil)
	// A failed edit undoes the pin of the same request (#39).
	c.do("PATCH", path, `{"pinned":true,"tags":["a b"]}`, http.StatusBadRequest, nil)
	if strings.Contains(c.raw("/api/pinned"), note.ID) {
		t.Errorf("a failed PATCH pinned the note: %s", c.raw("/api/pinned"))
	}

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
	c.do("PATCH", path, `{"pinned":true}`, http.StatusOK, &read)
	if !read.Pinned || read.Content != "changed" || !strings.Contains(c.raw(path), `"pinned":true`) || !strings.Contains(c.raw("/api/pinned"), note.ID) {
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

	// Its children come by pages from /api/items, not with it.
	if strings.Contains(c.raw(path), "children") {
		t.Errorf("a folder has a children field: %s", c.raw(path))
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
	c.do("PATCH", path, `{"pinned":false,"title":""}`, http.StatusBadRequest, nil)
	if !strings.Contains(c.raw("/api/pinned"), folder.ID) {
		t.Errorf("a failed PATCH removed the favorite: %s", c.raw("/api/pinned"))
	}

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
	if read.Path != "/"+title {
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

	// A page keeps the order, the total counts every child.
	c.do("GET", "/api/items?parent="+folder.ID, "", http.StatusOK, &items)
	r := httptest.NewRequest(http.MethodGet, "/api/items?limit=1&offset=1&parent="+folder.ID, nil)
	r.Host = "localhost:8080"
	w := httptest.NewRecorder()
	c.handler.ServeHTTP(w, r)
	var page []item
	if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil || len(page) != 1 || page[0].ID != items[1].ID || w.Header().Get("X-Total-Count") != "2" {
		t.Errorf("page = %+v, total %q, %v; want the note, total 2", page, w.Header().Get("X-Total-Count"), err)
	}
	if got := c.raw("/api/items?limit=1&offset=2&parent=" + folder.ID); strings.TrimSpace(got) != "[]" {
		t.Errorf("past the end = %q, want []", got)
	}
	c.do("GET", "/api/items?limit=-1", "", http.StatusBadRequest, nil)
	c.do("GET", "/api/items?offset=many", "", http.StatusBadRequest, nil)

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

	// A page of the notes keeps their order, the total counts them all (#143).
	r := httptest.NewRequest(http.MethodGet, "/api/tags/"+first+"?limit=1&offset=1", nil)
	r.Host = "localhost:8080"
	w := httptest.NewRecorder()
	c.handler.ServeHTTP(w, r)
	both := tag.Notes
	if err := json.Unmarshal(w.Body.Bytes(), &tag); err != nil || len(tag.Notes) != 1 || tag.Notes[0].ID != both[1].ID || w.Header().Get("X-Total-Count") != "2" {
		t.Errorf("page = %+v, total %q, %v; want the second note, total 2", tag.Notes, w.Header().Get("X-Total-Count"), err)
	}
	c.do("GET", "/api/tags/"+first+"?limit=-1", "", http.StatusBadRequest, nil)

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

	type laidOut struct {
		Notes []struct {
			ID   string   `json:"id"`
			Tags []string `json:"tags"`
			X    float64  `json:"x"`
			Y    float64  `json:"y"`
		} `json:"notes"`
		Links        []struct{}                    `json:"links"`
		TagPositions map[string]struct{ X, Y any } `json:"tag_positions"`
	}
	var notes laidOut
	c.do("GET", "/api/graph/notes", "", http.StatusOK, &notes)
	tagged := 0
	for _, note := range notes.Notes {
		if (note.ID == a.ID || note.ID == b.ID) && slices.Contains(note.Tags, first) && note.X > 0 && note.Y > 0 {
			tagged++
		}
	}
	if _, ok := notes.TagPositions[first]; tagged != 2 || notes.Links == nil || !ok {
		t.Errorf("graph notes: a and b carry %s laid out %d times, want 2; links %v; tag laid out %v", first, tagged, notes.Links, ok)
	}
	var near laidOut
	c.do("GET", "/api/notes/"+a.ID+"/graph", "", http.StatusOK, &near)
	if _, ok := near.TagPositions[first]; len(near.Notes) != 2 || near.Notes[0].ID != a.ID || near.Notes[0].X == 0 || !ok {
		t.Errorf("graph of a = %+v, want a and b laid out with %s", near, first)
	}
	c.do("GET", "/api/notes/"+unknownID+"/graph", "", http.StatusNotFound, nil)
	var stranger laidOut
	c.other(t).do("GET", "/api/graph/notes", "", http.StatusOK, &stranger)
	if len(stranger.Notes) != 0 {
		t.Errorf("another account sees %d notes in its graph", len(stranger.Notes))
	}
	c.other(t).do("GET", "/api/notes/"+a.ID+"/graph", "", http.StatusNotFound, nil)
	var list []struct {
		ID string `json:"id"`
	}
	c.do("GET", "/api/notes", "", http.StatusOK, &list)
	if len(list) != len(notes.Notes) || !strings.Contains(c.raw("/api/notes"), a.ID) {
		t.Errorf("notes = %d, want the %d of the graph with a", len(list), len(notes.Notes))
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
	if got := c.raw("/api/search?limit=1&offset=1&q=needle-" + first); strings.TrimSpace(got) != "[]" {
		t.Errorf("past the end = %q, want []", got)
	}
	c.do("GET", "/api/search?offset=many&q=needle", "", http.StatusBadRequest, nil)

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

// The browser sends loose files or a picked folder as a form, each file named by its path.
func TestImportForm(t *testing.T) {
	c := newClient(t)
	folder := fmt.Sprintf("form-%d", time.Now().UnixNano())
	form := func(names ...string) (string, string) {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		for _, name := range names {
			part, _ := writer.CreateFormFile("files", name)
			part.Write([]byte("# " + name))
		}
		writer.Close()
		return body.String(), writer.FormDataContentType()
	}
	send := func(body, kind string, status int, dest any) {
		t.Helper()
		r := httptest.NewRequest(http.MethodPost, "/api/import", strings.NewReader(body))
		r.Host = "localhost:8080"
		r.Header.Set("Content-Type", kind)
		w := httptest.NewRecorder()
		c.handler.ServeHTTP(w, r)
		if w.Code != status {
			t.Fatalf("POST /api/import: status = %d, want %d: %s", w.Code, status, w.Body)
		}
		if dest != nil {
			json.Unmarshal(w.Body.Bytes(), dest)
		}
	}

	var result struct{ Folders, Notes int }
	body, kind := form(folder+"/Un.md", folder+"/sous/Deux.md", folder+"-Trois.md")
	send(body, kind, http.StatusOK, &result)
	if result.Folders != 2 || result.Notes != 3 {
		t.Errorf("import = %+v, want 2 folders and 3 notes", result)
	}
	body, kind = form("../" + folder + ".md")
	send(body, kind, http.StatusBadRequest, nil)
	body, kind = form(folder+"-Double.md", folder+"-Double.md")
	send(body, kind, http.StatusBadRequest, nil)
	body, kind = form(folder + "/Coupé.md")
	send(body[:len(body)-10], kind, http.StatusBadRequest, nil)
}

func TestRouterGuards(t *testing.T) {
	handler := New(nil, nil, "http://localhost:3001", domain.AddUserModes[0])

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

	// An attachment downloads under its name.
	var pdf struct{ ID string }
	c.do(http.MethodPost, "/api/files?name=rapport.pdf", "%PDF-1.7 ...", http.StatusCreated, &pdf)
	r := httptest.NewRequest(http.MethodGet, "/api/files/"+pdf.ID, nil)
	r.Host = "localhost:8080"
	w := httptest.NewRecorder()
	c.handler.ServeHTTP(w, r)
	if got := w.Header().Get("Content-Disposition"); got != "attachment; filename=rapport.pdf" {
		t.Errorf("Content-Disposition = %q", got)
	}
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

	// A folder exported alone keeps its subfolders, and the images they cite.
	var top, sub item
	c.do(http.MethodPost, "/api/folders", fmt.Sprintf(`{"title":%q}`, unique("haut")), http.StatusCreated, &top)
	c.do(http.MethodPost, "/api/folders", fmt.Sprintf(`{"title":"bas","parent_id":%q}`, top.ID), http.StatusCreated, &sub)
	c.do(http.MethodPost, "/api/notes", fmt.Sprintf(`{"content":"# Chat\n\n![chat](/api/files/%s)","parent_id":%q}`, file.ID, sub.ID), http.StatusCreated, nil)
	exported = c.raw("/api/export?id=" + top.ID)
	read, err = zip.NewReader(strings.NewReader(exported), int64(len(exported)))
	if err != nil {
		t.Fatalf("export of a folder is not a zip: %v", err)
	}
	var names []string
	for _, entry := range read.File {
		names = append(names, entry.Name)
	}
	if want := []string{top.Title + "/", top.Title + "/bas/", top.Title + "/bas/Chat.md", "_files/", "_files/" + file.ID + ".png"}; !slices.Equal(names, want) {
		t.Errorf("export of a folder = %q, want %q", names, want)
	}
	if entry, err := read.Open(top.Title + "/bas/Chat.md"); err == nil {
		data, _ := io.ReadAll(entry)
		entry.Close()
		if want := "](../../_files/" + file.ID + ".png)"; !strings.Contains(string(data), want) {
			t.Errorf("Chat.md = %q, want %q in it", data, want)
		}
	}
	c.do(http.MethodGet, "/api/export?id=nope", "", http.StatusBadRequest, nil)
	c.do(http.MethodGet, "/api/export?id="+unknownID, "", http.StatusNotFound, nil)
	c.other(t).do(http.MethodGet, "/api/export?id="+top.ID, "", http.StatusNotFound, nil)

	// An image gone from Silo cuts the export off rather than ending a broken zip as a finished download.
	clear(c.store.Objects)
	defer func() {
		if got := recover(); got != http.ErrAbortHandler {
			t.Errorf("export without its image: recover() = %v, want http.ErrAbortHandler", got)
		}
	}()
	c.raw("/api/export")
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
	bob.do("GET", "/api/notes/"+note.ID+"/versions", "", http.StatusNotFound, nil)
	alice.do("PATCH", "/api/notes/"+note.ID, `{"content":"# Secret #privé v2"}`, http.StatusOK, nil)
	var versions []item
	alice.do("GET", "/api/notes/"+note.ID+"/versions", "", http.StatusOK, &versions)
	bob.do("GET", "/api/notes/"+note.ID+"/versions/"+versions[0].ID, "", http.StatusNotFound, nil)
	bob.do("DELETE", "/api/notes/"+note.ID, "", http.StatusNotFound, nil)
	bob.do("POST", "/api/notes", fmt.Sprintf(`{"content":"# Intrus","parent_id":%q}`, folder.ID), http.StatusBadRequest, nil)
	for _, path := range []string{"/api/items", "/api/search?q=Secret", "/api/graph", "/api/graph/notes", "/api/notes", "/api/links/broken", "/api/files", "/api/recent", "/api/folders", "/api/tags"} {
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
	if w := send("PATCH", "/api/me", `{"theme":"dracula","color_mode":"dark","view":"grid","plan":"pro","trash_retention_days":7}`, cookies[0]); w.Code != http.StatusOK ||
		!strings.Contains(w.Body.String(), `"locale":"en","theme":"dracula","color_mode":"dark","view":"grid","plan":"pro","admin":false,"trash_retention_days":7`) {
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

	login := func() *http.Cookie {
		t.Helper()
		cookies := send("POST", "/api/login", fmt.Sprintf(`{"name":%q,"password":"battery staple"}`, name), nil).Result().Cookies()
		if len(cookies) != 1 {
			t.Fatalf("login: cookies = %v", cookies)
		}
		return cookies[0]
	}

	// An expired session is refused.
	expired := login()
	if _, err := c.db.NewUpdate().Model((*domain.Session)(nil)).
		Set("expires_at = current_timestamp - interval '1 minute'").
		Where("token_hash = ?", password.TokenHash(expired.Value)).
		Exec(context.Background()); err != nil {
		t.Fatal(err)
	}
	if w := send("GET", "/api/me", "", expired); w.Code != http.StatusUnauthorized {
		t.Errorf("expired session: status = %d, want 401", w.Code)
	}

	// The admin's new password logs every device out.
	revoked := login()
	reset := resetpassword.Process(context.Background(), c.db, name, "horse battery")
	reset.Run()
	if reset.HasErr() {
		t.Fatal(reset.GetErr())
	}
	if w := send("GET", "/api/me", "", revoked); w.Code != http.StatusUnauthorized {
		t.Errorf("session after resetpassword: status = %d, want 401", w.Code)
	}
}

// account is an account as the admin panel lists it.
type account struct {
	ID, Name string
	Admin    bool
}

func TestAccounts(t *testing.T) {
	c := newClient(t)
	me := domain.UserID(c.ctx)

	send := func(method, path, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Host = "localhost:8080"
		w := httptest.NewRecorder()
		c.api.ServeHTTP(w, r)
		return w
	}
	admin := func(is bool) {
		t.Helper()
		if _, err := c.db.NewUpdate().Model((*domain.User)(nil)).Set("admin = ?", is).Where("id = ?", me).Exec(c.ctx); err != nil {
			t.Fatal(err)
		}
	}

	// Without the admin flag, the account manages nothing.
	admin(false)
	for _, route := range []string{"GET /api/users", "GET /api/invites", "POST /api/invites", "GET /api/version", "POST /api/users/" + me + "/password"} {
		method, path, _ := strings.Cut(route, " ")
		c.do(method, path, `{"name":"nope","password":"a-good-password"}`, http.StatusForbidden, nil)
	}

	admin(true)
	var accounts []account
	c.do("GET", "/api/users", "", http.StatusOK, &accounts)
	if !slices.ContainsFunc(accounts, func(a account) bool { return a.ID == me && a.Admin }) {
		t.Errorf("accounts = %+v, want this admin in it", accounts)
	}

	var version struct {
		Version string `json:"version"`
	}
	c.do("GET", "/api/version", "", http.StatusOK, &version)
	if version.Version == "" {
		t.Errorf("version = %+v, want a non-empty version", version)
	}

	// In the invitation mode, the sign-up needs an invitation: the base already has accounts.
	if w := send("GET", "/api/add-user", ""); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"allowed":false`) {
		t.Errorf("add-user without invitation: status = %d, body = %s", w.Code, w.Body)
	}
	invited := unique("invited")
	if w := send("POST", "/api/signup", fmt.Sprintf(`{"name":%q,"password":"a-good-password"}`, invited)); w.Code != http.StatusForbidden {
		t.Fatalf("signup without invitation: status = %d, want 403: %s", w.Code, w.Body)
	}

	var invite struct {
		Token  string
		Invite struct{ ID, Token string }
	}
	c.do("POST", "/api/invites", "", http.StatusCreated, &invite)
	if invite.Token == "" || invite.Invite.ID == "" || invite.Invite.Token != invite.Token {
		t.Fatalf("invite = %+v", invite)
	}
	if w := send("GET", "/api/add-user?invite="+invite.Token, ""); !strings.Contains(w.Body.String(), `"allowed":true`) {
		t.Errorf("add-user with the invitation: body = %s", w.Body)
	}

	// The base keeps the token sealed: the admin panel reads it back to show the link of a pending invitation again.
	var listed []struct{ ID, Token string }
	c.do("GET", "/api/invites", "", http.StatusOK, &listed)
	if !slices.ContainsFunc(listed, func(l struct{ ID, Token string }) bool {
		return l.ID == invite.Invite.ID && l.Token == invite.Token
	}) {
		t.Errorf("the invitation does not give its token back: %+v", listed)
	}

	// A name already taken does not spend the invitation, a password too short neither.
	if w := send("POST", "/api/signup", fmt.Sprintf(`{"name":%q,"password":"a-good-password","invite":%q}`, accounts[0].Name, invite.Token)); w.Code != http.StatusConflict {
		t.Errorf("a name taken: status = %d, want 409: %s", w.Code, w.Body)
	}
	if w := send("POST", "/api/signup", fmt.Sprintf(`{"name":%q,"password":"short","invite":%q}`, unique("short"), invite.Token)); w.Code != http.StatusBadRequest {
		t.Errorf("a short password: status = %d, want 400: %s", w.Code, w.Body)
	}

	// The invitation signs the account up and logs it in at once; it serves only once.
	w := send("POST", "/api/signup", fmt.Sprintf(`{"name":%q,"password":"a-good-password","invite":%q}`, invited, invite.Token))
	if cookies := w.Result().Cookies(); w.Code != http.StatusOK || len(cookies) != 1 || !cookies[0].HttpOnly {
		t.Fatalf("signup: status = %d, cookies = %v: %s", w.Code, w.Result().Cookies(), w.Body)
	}
	if w := send("POST", "/api/signup", fmt.Sprintf(`{"name":%q,"password":"a-good-password","invite":%q}`, unique("twice"), invite.Token)); w.Code != http.StatusForbidden {
		t.Errorf("the invitation used twice: status = %d, want 403: %s", w.Code, w.Body)
	}
	var pending []struct{ ID string }
	c.do("GET", "/api/invites", "", http.StatusOK, &pending)
	if slices.ContainsFunc(pending, func(p struct{ ID string }) bool { return p.ID == invite.Invite.ID }) {
		t.Errorf("the used invitation is still listed: %+v", pending)
	}

	// A revoked invitation signs nobody up.
	c.do("POST", "/api/invites", "", http.StatusCreated, &invite)
	c.do("DELETE", "/api/invites/"+invite.Invite.ID, "", http.StatusNoContent, nil)
	c.do("DELETE", "/api/invites/"+invite.Invite.ID, "", http.StatusNotFound, nil)
	if w := send("POST", "/api/signup", fmt.Sprintf(`{"name":%q,"password":"a-good-password","invite":%q}`, unique("revoked"), invite.Token)); w.Code != http.StatusForbidden {
		t.Errorf("a revoked invitation: status = %d, want 403: %s", w.Code, w.Body)
	}

	// An account is deleted with everything it holds, but never by itself.
	c.do("GET", "/api/users", "", http.StatusOK, &accounts)
	signedUp, found := slices.BinarySearchFunc(accounts, invited, func(a account, name string) int {
		return strings.Compare(a.Name, name)
	})
	if !found {
		t.Fatalf("the account signed up is not listed: %+v", accounts)
	}

	// The admin gives it a new password, shown once: the old one no longer logs in.
	var reset struct{ Password string }
	c.do("POST", "/api/users/"+unknownID+"/password", "", http.StatusNotFound, nil)
	c.do("POST", "/api/users/"+accounts[signedUp].ID+"/password", "", http.StatusOK, &reset)
	if w := send("POST", "/api/login", fmt.Sprintf(`{"name":%q,"password":"a-good-password"}`, invited)); w.Code != http.StatusUnauthorized {
		t.Errorf("the old password after a reset: status = %d, want 401: %s", w.Code, w.Body)
	}
	if w := send("POST", "/api/login", fmt.Sprintf(`{"name":%q,"password":%q}`, invited, reset.Password)); w.Code != http.StatusOK {
		t.Errorf("the new password after a reset: status = %d, want 200: %s", w.Code, w.Body)
	}

	c.do("DELETE", "/api/users/"+me, "", http.StatusConflict, nil)
	c.do("DELETE", "/api/users/"+unknownID, "", http.StatusNotFound, nil)
	c.do("DELETE", "/api/users/"+accounts[signedUp].ID, "", http.StatusNoContent, nil)
	c.do("GET", "/api/users", "", http.StatusOK, &accounts)
	if slices.ContainsFunc(accounts, func(a account) bool { return a.Name == invited }) {
		t.Errorf("the account deleted is still listed: %+v", accounts)
	}
}

func TestAccountsPublicMode(t *testing.T) {
	c := newClient(t)
	api := New(bun.IDB(c.db), c.store, "", domain.AddUserModes[1])

	send := func(path, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
		r.Host = "localhost:8080"
		w := httptest.NewRecorder()
		api.ServeHTTP(w, r)
		return w
	}

	// Anyone reaching the web app creates an account, no invitation needed.
	w := send("/api/signup", fmt.Sprintf(`{"name":%q,"password":"a-good-password"}`, unique("public")))
	if w.Code != http.StatusOK {
		t.Fatalf("signup in the public mode: status = %d, want 200: %s", w.Code, w.Body)
	}

	// One address creates a few accounts, then waits.
	for i := range 3 {
		if w := send("/api/signup", fmt.Sprintf(`{"name":%q,"password":"a-good-password"}`, unique("public"))); w.Code != http.StatusOK {
			t.Fatalf("signup %d in a row: status = %d, want 200: %s", i+2, w.Code, w.Body)
		}
	}
	if w := send("/api/signup", fmt.Sprintf(`{"name":%q,"password":"a-good-password"}`, unique("public"))); w.Code != http.StatusTooManyRequests {
		t.Errorf("signup 5 in a row: status = %d, want 429: %s", w.Code, w.Body)
	}
}
