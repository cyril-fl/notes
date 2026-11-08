package createfolder

import (
	"context"
	"errors"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
	"testing"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/deleteitems"
)

// No DB needed: makeFolder fails first and saveFolder is guarded, so a nil db is never touched.
func TestNoTitle(t *testing.T) {
	process := Process(context.Background(), nil, Input{Title: "   "})
	process.Run()

	if !errors.Is(process.GetErr(), ErrNoTitle) {
		t.Fatalf("err = %v, want %v", process.GetErr(), ErrNoTitle)
	}
}

func TestCreateFolder(t *testing.T) {
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

	root := Process(ctx, tx, Input{Title: " Root "})
	root.Run()
	if root.HasErr() {
		t.Fatalf("root: %v", root.GetErr())
	}
	if item := root.GetResult(); item.ID == "" || item.Title != "Root" || item.Kind != domain.FOLDER {
		t.Fatalf("result = %+v, want an id, Root, folder", item)
	}

	child := Process(ctx, tx, Input{Title: "Child", ParentID: root.GetResult().ID})
	child.Run()
	if child.HasErr() {
		t.Fatalf("child: %v", child.GetErr())
	}

	// A note is not a folder: the composite FK (parent_id, parent_kind) refuses it.
	note := createnote.Process(ctx, tx, createnote.Input{Content: "content"})
	note.Run()
	if note.HasErr() {
		t.Fatalf("note: %v", note.GetErr())
	}

	// A folder in the trash: the FK accepts it, the note would be lost with it.
	trash := deleteitems.Process(ctx, tx, domain.FOLDER, []string{child.GetResult().ID})
	trash.Run()
	if trash.HasErr() {
		t.Fatalf("trash: %v", trash.GetErr())
	}

	// Last: DB errors (RunInTx uses a savepoint here, but keep them at the end).
	for _, parentID := range []string{"00000000-0000-0000-0000-000000000000", note.GetResult().ID, child.GetResult().ID} {
		orphan := Process(ctx, tx, Input{Title: "Orphan", ParentID: parentID})
		orphan.Run()
		if !errors.Is(orphan.GetErr(), ErrParentNotFound) {
			t.Errorf("parent %s: err = %v, want %v", parentID, orphan.GetErr(), ErrParentNotFound)
		}
	}
}
