package readitem

import (
	"context"
	"errors"
	"testing"

	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createfolder"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/deleteitems"
)

// No DB needed: makeID fails first and the loads are guarded, so a nil db is never touched.
func TestNoValue(t *testing.T) {
	process := Process(context.Background(), nil, domain.NOTE, " ")
	process.Run()
	if !errors.Is(process.GetErr(), ErrNoValue) {
		t.Fatalf("err = %v, want %v", process.GetErr(), ErrNoValue)
	}
}

// The found case is covered by editnote's test, which reads the note back.
func TestNotFound(t *testing.T) {
	env := postgrey.NewEnvironment()
	env.ReadEnv()
	db := env.Connect()
	defer db.Close()

	for _, id := range []string{"00000000-0000-0000-0000-000000000000", "nope"} {
		process := Process(context.Background(), db, domain.FOLDER, id)
		process.Run()
		if !errors.Is(process.GetErr(), ErrNotFound) {
			t.Errorf("%s: err = %v, want %v", id, process.GetErr(), ErrNotFound)
		}
	}
}

func TestPathAndChildren(t *testing.T) {
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

	a := folder(t, tx, "A", "")
	b := folder(t, tx, "B", a)
	note := createnote.Process(ctx, tx, createnote.Input{Content: "# N", ParentID: a})
	note.Run()
	untitled := createnote.Process(ctx, tx, createnote.Input{Content: "```\ncode\n```", ParentID: b})
	untitled.Run()
	trashed := folder(t, tx, "Trashed", a)
	trash := deleteitems.Process(ctx, tx, domain.FOLDER, []string{trashed})
	trash.Run()
	if note.HasErr() || untitled.HasErr() || trash.HasErr() {
		t.Fatalf("seed: %v %v %v", note.GetErr(), untitled.GetErr(), trash.GetErr())
	}

	read := Process(ctx, tx, domain.NOTE, untitled.GetResult().ID)
	read.Run()
	if read.HasErr() || read.GetResult().Path != "/A/B/untitled" {
		t.Errorf("path = %q, %v; want /A/B/untitled", read.GetResult().Path, read.GetErr())
	}

	// Folders first, then notes; the trashed folder is gone.
	read = Process(ctx, tx, domain.FOLDER, a)
	read.Run()
	got := read.GetResult()
	if read.HasErr() || got.Path != "/A" || len(got.Children) != 2 || got.Children[0].ID != b || got.Children[1].ID != note.GetResult().ID {
		t.Errorf("folder = %+v, %v; want /A with [B, N]", got, read.GetErr())
	}
}

func folder(t *testing.T, db bun.IDB, title, parentID string) string {
	t.Helper()
	process := createfolder.Process(context.Background(), db, createfolder.Input{Title: title, ParentID: parentID})
	process.Run()
	if process.HasErr() {
		t.Fatalf("folder %s: %v", title, process.GetErr())
	}
	return process.GetResult().ID
}
