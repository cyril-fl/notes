package layoutgraph

import (
	"bytes"
	"context"
	"testing"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/forcelayout"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readgraph"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
	"github.com/uptrace/bun"
)

func TestLayoutGraph(t *testing.T) {
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
	// A tx keeps the ctx it began with: the other account's note before the tested one.
	other, tx := testaccount.New(t, ctx, tx)
	note(t, other, tx, "# Other", "secret")
	ctx, tx = testaccount.New(t, ctx, tx)

	ids := map[string]string{}
	ids["Target"] = note(t, ctx, tx, "# Target")
	ids["Source"] = note(t, ctx, tx, "# Source\n\n[[Target]]", "go")
	ids["Mate"] = note(t, ctx, tx, "# Mate", "go")
	ids["Far"] = note(t, ctx, tx, "# Far", "web")

	first := layout(t, ctx, tx, "")
	if len(first.Notes) != 4 {
		t.Fatalf("%d notes, want 4: the other account's are not here", len(first.Notes))
	}
	if _, ok := first.TagPositions["secret"]; ok || len(first.TagPositions) != 2 {
		t.Errorf("tag positions %v, want go and web", first.TagPositions)
	}
	for _, n := range first.Notes {
		if n.X < 8 || n.X > 92 || n.Y < 8 || n.Y > 92 {
			t.Errorf("%s at %v, %v: out of the view", n.Path, n.X, n.Y)
		}
		if len(n.TagStrengths) != len(n.Tags) {
			t.Errorf("%s: %d strengths for %d tags", n.Path, len(n.TagStrengths), len(n.Tags))
		}
	}
	if len(first.Links) != 1 || first.Links[0].Strength != 1 {
		t.Errorf("links %+v, want Source-Target at full strength", first.Links)
	}

	// Saved, sealed: another device gets the same map without computing it.
	var row domain.GraphLayout
	if err := tx.NewSelect().Model(&row).Where("user_id = ?", domain.UserID(ctx)).Scan(ctx); err != nil {
		t.Fatalf("layout not saved: %v", err)
	}
	if string(row.Data) == "" || containsText(row.Data, "web") {
		t.Error("the saved layout names the tags in clear")
	}
	again := layout(t, ctx, tx, "")
	if place(again, ids["Far"]) != place(first, ids["Far"]) {
		t.Error("the saved layout was computed again")
	}
	if count, err := tx.NewSelect().Model((*domain.GraphLayout)(nil)).Where("user_id = ?", domain.UserID(other)).Count(ctx); err != nil || count != 0 {
		t.Errorf("the other account got a layout: %d, %v", count, err)
	}

	// A new note: the graph changed, the layout is computed again and saved.
	ids["New"] = note(t, ctx, tx, "# New", "web")
	changed := layout(t, ctx, tx, "")
	if len(changed.Notes) != 5 || place(changed, ids["New"]) == (forcelayout.Point{}) {
		t.Errorf("the new note is not laid out: %+v", changed.Notes)
	}
	var after domain.GraphLayout
	if err := tx.NewSelect().Model(&after).Where("user_id = ?", domain.UserID(ctx)).Scan(ctx); err != nil || string(after.Data) == string(row.Data) {
		t.Errorf("the changed layout was not saved: %v", err)
	}

	// The neighborhood of Source: its tag, the note sharing it, the note of its link.
	near := layout(t, ctx, tx, ids["Source"])
	got := map[string]bool{}
	for _, n := range near.Notes {
		got[n.ID] = true
	}
	if len(near.Notes) != 3 || !got[ids["Source"]] || !got[ids["Mate"]] || !got[ids["Target"]] {
		t.Errorf("neighborhood %+v, want Source, Mate and Target", near.Notes)
	}
	if len(near.Links) != 1 || near.Links[0].Source != ids["Source"] || near.Links[0].Target != ids["Target"] {
		t.Errorf("neighborhood links %+v", near.Links)
	}
	if _, ok := near.TagPositions["go"]; !ok {
		t.Error("the tag go is not laid out")
	}
}

// Helpers

func note(t *testing.T, ctx context.Context, tx bun.IDB, content string, tags ...string) string {
	t.Helper()
	process := createnote.Process(ctx, tx, createnote.Input{Content: content, Tags: tags})
	process.Run()
	if process.HasErr() {
		t.Fatalf("note %q: %v", content, process.GetErr())
	}
	return process.GetResult().ID
}

func layout(t *testing.T, ctx context.Context, tx bun.IDB, noteID string) readgraph.Result {
	t.Helper()
	process := Process(ctx, tx, noteID)
	process.Run()
	if process.HasErr() {
		t.Fatalf("layout %q: %v", noteID, process.GetErr())
	}
	return process.GetResult()
}

func place(result readgraph.Result, id string) forcelayout.Point {
	for _, n := range result.Notes {
		if n.ID == id {
			return forcelayout.Point{X: n.X, Y: n.Y}
		}
	}
	return forcelayout.Point{}
}

func containsText(data []byte, text string) bool {
	return bytes.Contains(data, []byte(text))
}
