package emptytrash

import (
	"context"
	"errors"
	"main.go/app/shared/services/testaccount"
	"testing"

	"github.com/uptrace/bun"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/deleteitems"
)

func TestEmptyTrashOfIDs(t *testing.T) {
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

	var ids []string
	for _, content := range []string{"# Purged", "# Kept in the trash", "# Live"} {
		note := createnote.Process(ctx, tx, createnote.Input{Content: content})
		note.Run()
		if note.HasErr() {
			t.Fatalf("note: %v", note.GetErr())
		}
		ids = append(ids, note.GetResult().ID)
	}
	trash := deleteitems.Process(ctx, tx, domain.NOTE, ids[:2])
	trash.Run()
	if trash.HasErr() {
		t.Fatalf("trash: %v", trash.GetErr())
	}

	purge := Process(ctx, tx, ids[0])
	purge.Run()
	if purge.HasErr() {
		t.Fatalf("purge: %v", purge.GetErr())
	}

	var left []string
	if err := tx.NewRaw("SELECT id FROM items WHERE id IN (?)", bun.List(ids)).Scan(ctx, &left); err != nil {
		t.Fatal(err)
	}
	if len(left) != 2 || left[0] == ids[0] || left[1] == ids[0] {
		t.Errorf("left = %v, want all but %s", left, ids[0])
	}

	// A live note, or one already gone, is not in the trash.
	for _, id := range []string{ids[2], ids[0], "nope"} {
		missing := Process(ctx, tx, id)
		missing.Run()
		if !errors.Is(missing.GetErr(), ErrNotFound) {
			t.Errorf("purge %q: err = %v, want %v", id, missing.GetErr(), ErrNotFound)
		}
	}
}
