package readneighbors

import (
	"context"
	"errors"
	"fmt"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/deleteitems"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
)

func TestSnippet(t *testing.T) {
	id, other := "0a6e3f1c-2b4d-4e5f-8a9b-0c1d2e3f4a5b", "11111111-2222-4333-8444-555555555555"
	long := strings.Repeat("a", 80) + " [Center](note:" + id + ") " + strings.Repeat("b", 80)
	tests := map[string]struct{ content, want string }{
		"its line":         {"# Source\n\nsee [Center](note:" + strings.ToUpper(id) + ") here", "see [[Center]] here"},
		"other links read": {"[Other](note:" + other + ") then [Center#Part](note:" + id + "#Part)", "[[Other]] then [[Center#Part]]"},
		"cut around":       {long, "…" + strings.Repeat("a", 19) + " [[Center]] " + strings.Repeat("b", 29) + "…"},
		"no link":          {"# Source\n\nnothing", ""},
	}

	for name, test := range tests {
		if got := snippet(test.content, id); got != test.want {
			t.Errorf("%s: snippet() = %q, want %q", name, got, test.want)
		}
	}
}

func TestNeighbors(t *testing.T) {
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

	// Unique per run so real data in the DB cannot collide.
	a, b, c := fmt.Sprintf("test-%d-a", time.Now().UnixNano()), fmt.Sprintf("test-%d-b", time.Now().UnixNano()), fmt.Sprintf("test-%d-c", time.Now().UnixNano())
	ids := map[string]string{}
	for _, note := range []struct {
		title string
		tags  []string
	}{
		{"Center", []string{a, b}},
		{"Two", []string{a, b, c}},
		{"One", []string{b}},
		{"None", []string{c}},
		{"Trashed", []string{a}},
	} {
		process := createnote.Process(ctx, tx, createnote.Input{Content: "# " + note.title, Tags: note.tags})
		process.Run()
		if process.HasErr() {
			t.Fatalf("note %s: %v", note.title, process.GetErr())
		}
		ids[note.title] = process.GetResult().ID
	}
	trash := deleteitems.Process(ctx, tx, domain.NOTE, []string{ids["Trashed"]})
	trash.Run()
	if trash.HasErr() {
		t.Fatalf("trash: %v", trash.GetErr())
	}

	// Two shares a and b, One shares b; None shares nothing, Trashed is in the trash.
	process := Process(ctx, tx, ids["Center"])
	process.Run()
	got := process.GetResult()
	if process.HasErr() || got.Note.Item.Title != "Center" || len(got.Neighbors) != 2 {
		t.Fatalf("result = %+v, %v; want Center with 2 neighbors", got, process.GetErr())
	}
	if two := got.Neighbors[0]; two.ID != ids["Two"] || two.Shared != 2 || !slices.Equal(two.Tags, []string{a, b}) || two.Path != "/Two" {
		t.Errorf("first neighbor = %+v, want Two sharing [a b]", two)
	}
	if one := got.Neighbors[1]; one.ID != ids["One"] || one.Shared != 1 {
		t.Errorf("second neighbor = %+v, want One sharing 1", one)
	}

	// A note citing Center is a backlink with the line of the link, shown with its label.
	source := createnote.Process(ctx, tx, createnote.Input{Content: "# Source\n\nsee [[Center]] here"})
	source.Run()
	if source.HasErr() {
		t.Fatalf("source: %v", source.GetErr())
	}
	process = Process(ctx, tx, ids["Center"])
	process.Run()
	if back := process.GetResult().Backlinks; process.HasErr() || len(back) != 1 || back[0].Snippet != "see [[Center]] here" {
		t.Errorf("backlinks = %+v, %v; want Source with its line", back, process.GetErr())
	}

	missing := Process(ctx, tx, "00000000-0000-0000-0000-000000000000")
	missing.Run()
	if !errors.Is(missing.GetErr(), readitem.ErrNotFound) {
		t.Errorf("err = %v, want %v", missing.GetErr(), readitem.ErrNotFound)
	}
}
