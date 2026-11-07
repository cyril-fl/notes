package pinnote

import (
	"context"
	"errors"
	"fmt"
	"main.go/app/shared/services/testaccount"
	"slices"
	"testing"
	"time"

	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createfolder"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/deleteitems"
	"main.go/app/shared/services/process/listpinned"
)

func TestPin(t *testing.T) {
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
	for _, content := range []string{"# Pinned", "# Trashed"} {
		note := createnote.Process(ctx, tx, createnote.Input{Content: content})
		note.Run()
		if note.HasErr() {
			t.Fatalf("note: %v", note.GetErr())
		}
		ids = append(ids, note.GetResult().ID)
	}
	for _, id := range ids {
		pin := Process(ctx, tx, domain.NOTE, id, true)
		pin.Run()
		if pin.HasErr() {
			t.Fatalf("pin: %v", pin.GetErr())
		}
	}
	trash := deleteitems.Process(ctx, tx, domain.NOTE, ids[1:])
	trash.Run()
	if trash.HasErr() {
		t.Fatalf("trash: %v", trash.GetErr())
	}

	// Pinning is not an edit, and a trashed note leaves the list.
	var updated time.Time
	if err := tx.NewSelect().Model((*domain.Item)(nil)).Column("updated_at").Where("id = ?", ids[0]).Scan(ctx, &updated); err != nil || !updated.IsZero() {
		t.Errorf("updated_at = %v, %v; want unset", updated, err)
	}
	if got := pinned(t, ctx, tx); !slices.Contains(got, ids[0]) || slices.Contains(got, ids[1]) {
		t.Errorf("pinned = %v, want the live note only", got)
	}

	unpin := Process(ctx, tx, domain.NOTE, ids[0], false)
	unpin.Run()
	if unpin.HasErr() {
		t.Fatalf("unpin: %v", unpin.GetErr())
	}
	if got := pinned(t, ctx, tx); slices.Contains(got, ids[0]) {
		t.Errorf("pinned = %v, want the note unpinned", got)
	}

	// A folder is a favorite the same way, but only as a folder.
	folder := createfolder.Process(ctx, tx, createfolder.Input{Title: fmt.Sprintf("fav-%d", time.Now().UnixNano())})
	folder.Run()
	if folder.HasErr() {
		t.Fatalf("folder: %v", folder.GetErr())
	}
	fav := Process(ctx, tx, domain.FOLDER, folder.GetResult().ID, true)
	fav.Run()
	if fav.HasErr() {
		t.Fatalf("fav: %v", fav.GetErr())
	}
	if got := pinned(t, ctx, tx); !slices.Contains(got, folder.GetResult().ID) {
		t.Errorf("pinned = %v, want the favorite folder", got)
	}
	wrong := Process(ctx, tx, domain.NOTE, folder.GetResult().ID, true)
	wrong.Run()
	if !errors.Is(wrong.GetErr(), ErrNotFound) {
		t.Errorf("pin a folder as a note: err = %v, want %v", wrong.GetErr(), ErrNotFound)
	}

	for _, id := range []string{ids[1], "00000000-0000-0000-0000-000000000000", "nope"} {
		missing := Process(ctx, tx, domain.NOTE, id, true)
		missing.Run()
		if !errors.Is(missing.GetErr(), ErrNotFound) {
			t.Errorf("pin %q: err = %v, want %v", id, missing.GetErr(), ErrNotFound)
		}
	}
}

func pinned(t *testing.T, ctx context.Context, tx bun.Tx) []string {
	t.Helper()
	list := listpinned.Process(ctx, tx)
	list.Run()
	if list.HasErr() {
		t.Fatalf("list: %v", list.GetErr())
	}
	var ids []string
	for _, note := range list.GetResult() {
		ids = append(ids, note.ID)
	}
	return ids
}
