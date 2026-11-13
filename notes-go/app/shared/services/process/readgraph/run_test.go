package readgraph

import (
	"context"
	"fmt"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
	"slices"
	"testing"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createfolder"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/deleteitems"
)

func TestGraph(t *testing.T) {
	ctx := context.Background()
	env := postgrey.NewEnvironment()
	env.ReadEnv()
	db := env.Connect()
	defer db.Close()

	schema := createschema.Process(ctx, db)
	schema.Run()
	if schema.HasErr() {
		t.Fatalf("create schema: %v", schema.GetErr())
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	ctx, tx = testaccount.New(t, ctx, tx)

	folder := createfolder.Process(ctx, tx, createfolder.Input{Title: "Gone"})
	folder.Run()
	if folder.HasErr() {
		t.Fatalf("folder: %v", folder.GetErr())
	}

	// Unique per run so real data in the DB cannot collide.
	a, b := fmt.Sprintf("test-%d-a", time.Now().UnixNano()), fmt.Sprintf("test-%d-b", time.Now().UnixNano())
	ids := map[string]string{}
	for _, note := range []struct {
		title    string
		parentID string
		tags     []string
	}{
		{"Two", "", []string{a, b}},
		{"Three", "", []string{a, b}},
		{"One", "", []string{b}},
		{"Alone", "", nil},
		{"Trashed", "", []string{a}},
		{"InTrashedFolder", folder.GetResult().ID, []string{a}},
	} {
		process := createnote.Process(ctx, tx, createnote.Input{Content: "# " + note.title, ParentID: note.parentID, Tags: note.tags})
		process.Run()
		if process.HasErr() {
			t.Fatalf("note %s: %v", note.title, process.GetErr())
		}
		ids[note.title] = process.GetResult().ID
	}
	// A [[link]] joins two notes sharing no tag.
	for _, note := range []struct{ key, content string }{{"Target", "# " + a + "-target"}, {"Source", "# " + a + "-source\n\n[[" + a + "-target]]"}} {
		process := createnote.Process(ctx, tx, createnote.Input{Content: note.content})
		process.Run()
		if process.HasErr() {
			t.Fatalf("note %s: %v", note.key, process.GetErr())
		}
		ids[note.key] = process.GetResult().ID
	}
	for kind, id := range map[domain.Kind]string{domain.NOTE: ids["Trashed"], domain.FOLDER: folder.GetResult().ID} {
		trash := deleteitems.Process(ctx, tx, kind, []string{id})
		trash.Run()
		if trash.HasErr() {
			t.Fatalf("trash: %v", trash.GetErr())
		}
	}

	process := Process(ctx, tx, true)
	process.Run()
	if process.HasErr() {
		t.Fatalf("graph: %v", process.GetErr())
	}
	got := process.GetResult()

	for _, title := range []string{"Two", "Three", "One", "Alone"} {
		if !hasNote(got, ids[title]) {
			t.Errorf("note %s is missing", title)
		}
	}
	for _, title := range []string{"Trashed", "InTrashedFolder"} {
		if hasNote(got, ids[title]) {
			t.Errorf("note %s should be hidden", title)
		}
	}

	if tags := noteTags(got, ids["Two"]); !slices.Equal(tags, []string{a, b}) {
		t.Errorf("Two tags = %v, want [a b]", tags)
	}
	if tags := noteTags(got, ids["One"]); !slices.Equal(tags, []string{b}) {
		t.Errorf("One tags = %v, want [b]", tags)
	}
	// Only the [[links]]: sharing tags makes no link.
	if len(got.Links) != 1 || findLink(got, ids["Source"], ids["Target"]) == nil {
		t.Errorf("links = %+v, want Source-Target alone", got.Links)
	}

	notes := Process(ctx, tx, false)
	notes.Run()
	if notes.HasErr() {
		t.Fatalf("notes: %v", notes.GetErr())
	}
	if only := notes.GetResult(); len(only.Notes) != len(got.Notes) || only.Links != nil || noteTags(only, ids["Two"]) != nil {
		t.Errorf("without links = %d notes, %d links, want the %d notes alone", len(only.Notes), len(only.Links), len(got.Notes))
	}
}

// Helpers

func hasNote(result Result, id string) bool {
	return slices.ContainsFunc(result.Notes, func(note Note) bool { return note.ID == id })
}

func findLink(result Result, a, b string) *Link {
	for i, link := range result.Links {
		if link.Source == a && link.Target == b || link.Source == b && link.Target == a {
			return &result.Links[i]
		}
	}
	return nil
}

func noteTags(result Result, id string) []string {
	for _, note := range result.Notes {
		if note.ID == id {
			return note.Tags
		}
	}
	return nil
}
