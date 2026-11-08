package restoreitems

import (
	"context"
	"errors"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
	"testing"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createfolder"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/deleteitems"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listtrash"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
)

// No DB needed: makeIDs fails first and restoreItems is guarded, so a nil db is never touched.
func TestNoValue(t *testing.T) {
	process := Process(context.Background(), nil, []string{" "})
	process.Run()
	if !errors.Is(process.GetErr(), ErrNoValue) {
		t.Fatalf("err = %v, want %v", process.GetErr(), ErrNoValue)
	}
}

// Covers listtrash too: it reads the trash these removals build.
func TestTrashAndRestore(t *testing.T) {
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

	folder := func(title, parentID string) string {
		process := createfolder.Process(ctx, tx, createfolder.Input{Title: title, ParentID: parentID})
		process.Run()
		if process.HasErr() {
			t.Fatalf("folder %s: %v", title, process.GetErr())
		}
		return process.GetResult().ID
	}
	note := func(title, parentID string) string {
		process := createnote.Process(ctx, tx, createnote.Input{Content: "# " + title, ParentID: parentID})
		process.Run()
		if process.HasErr() {
			t.Fatalf("note %s: %v", title, process.GetErr())
		}
		return process.GetResult().ID
	}
	remove := func(kind domain.Kind, id string) {
		process := deleteitems.Process(ctx, tx, kind, []string{id})
		process.Run()
		if process.HasErr() {
			t.Fatalf("remove %s: %v", id, process.GetErr())
		}
	}

	// Box holds Sub (with In inside) and Early; Early was removed a day before Box.
	box := folder("Box", "")
	sub := folder("Sub", box)
	in := note("In", sub)
	early := note("Early", box)
	remove(domain.NOTE, early)
	if _, err := tx.NewRaw("UPDATE items SET deleted_at = deleted_at - interval '1 day' WHERE id = ?", early).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	remove(domain.FOLDER, box)

	trash := listtrash.Process(ctx, tx)
	trash.Run()
	roots := map[string]listtrash.Trashed{}
	for _, item := range trash.GetResult() {
		roots[item.ID] = item
	}
	if trash.HasErr() || roots[box].Inside != 2 || roots[early].Title != "Early" {
		t.Fatalf("trash = %+v, %v; want Box with 2 inside, and Early on its own", roots, trash.GetErr())
	}
	if _, ok := roots[sub]; ok {
		t.Error("Sub listed on its own, want it inside Box")
	}

	// Early's parent Box is still in the trash.
	blocked := Process(ctx, tx, []string{early})
	blocked.Run()
	if !errors.Is(blocked.GetErr(), ErrParentInTrash) {
		t.Fatalf("err = %v, want %v", blocked.GetErr(), ErrParentInTrash)
	}

	// Box comes back with Sub and In; Early, removed earlier, stays; a live id is skipped.
	restore := Process(ctx, tx, []string{box, note("Live", "")})
	restore.Run()
	restored := map[string]bool{}
	for _, item := range restore.GetResult() {
		restored[item.ID] = true
	}
	if restore.HasErr() || len(restored) != 3 || !restored[box] || !restored[sub] || !restored[in] {
		t.Errorf("restored = %v, %v; want Box, Sub, In", restored, restore.GetErr())
	}

	// Now Early can come back.
	again := Process(ctx, tx, []string{early})
	again.Run()
	if again.HasErr() || len(again.GetResult()) != 1 {
		t.Errorf("early = %+v, %v; want it restored", again.GetResult(), again.GetErr())
	}

	// Last: a DB error.
	invalid := Process(ctx, tx, []string{"nope"})
	invalid.Run()
	if !errors.Is(invalid.GetErr(), ErrInvalidID) {
		t.Errorf("err = %v, want %v", invalid.GetErr(), ErrInvalidID)
	}
}

// A [[link]] waiting for a trashed note resolves when the note comes back.
func TestRestoreLinks(t *testing.T) {
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

	note := func(content string) string {
		process := createnote.Process(ctx, tx, createnote.Input{Content: content})
		process.Run()
		if process.HasErr() {
			t.Fatalf("note %q: %v", content, process.GetErr())
		}
		return process.GetResult().ID
	}

	b := note("# Bee")
	remove := deleteitems.Process(ctx, tx, domain.NOTE, []string{b})
	remove.Run()
	if remove.HasErr() {
		t.Fatal(remove.GetErr())
	}
	a := note("see [[Bee]]")

	restore := Process(ctx, tx, []string{b})
	restore.Run()
	if restore.HasErr() {
		t.Fatal(restore.GetErr())
	}

	var content secret.Text
	var links int
	if err := tx.NewRaw("SELECT content FROM metadata_notes WHERE item_id = ?", a).Scan(ctx, &content); err != nil {
		t.Fatal(err)
	}
	if err := tx.NewRaw("SELECT count(*) FROM notes_on_links WHERE source_id = ? AND target_id = ?", a, b).Scan(ctx, &links); err != nil {
		t.Fatal(err)
	}
	if want := "see [Bee](note:" + b + ")"; string(content) != want || links != 1 {
		t.Errorf("content = %q, links = %d; want %q, 1", content, links, want)
	}
}
