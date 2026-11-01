package moveitems

import (
	"context"
	"errors"
	"testing"

	"github.com/uptrace/bun"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createfolder"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
)

const unknown = "00000000-0000-0000-0000-000000000000"

// No DB needed: makeInput fails first and moveItems is guarded, so a nil db is never touched.
func TestNoValue(t *testing.T) {
	process := Process(context.Background(), nil, Input{IDs: []string{" ", ""}})
	process.Run()
	if !errors.Is(process.GetErr(), ErrNoValue) {
		t.Fatalf("err = %v, want %v", process.GetErr(), ErrNoValue)
	}
}

func TestMoveItems(t *testing.T) {
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
	b := folder(t, tx, "B", "")
	child := folder(t, tx, "Child", b)
	note := createnote.Process(ctx, tx, createnote.Input{Content: "C"})
	note.Run()
	if note.HasErr() {
		t.Fatalf("note: %v", note.GetErr())
	}

	// Note and folder into A; the unknown id is skipped.
	into := Process(ctx, tx, Input{IDs: []string{note.GetResult().ID, b, unknown}, ParentID: a})
	into.Run()
	if into.HasErr() || len(into.GetResult()) != 2 {
		t.Fatalf("into A = %+v, %v; want 2 items", into.GetResult(), into.GetErr())
	}
	for _, item := range into.GetResult() {
		if item.ParentID == nil || *item.ParentID != a {
			t.Errorf("%s parent = %v, want %s", item.ID, item.ParentID, a)
		}
	}

	// Back to the root: no parent.
	root := Process(ctx, tx, Input{IDs: []string{b}})
	root.Run()
	if root.HasErr() || len(root.GetResult()) != 1 || root.GetResult()[0].ParentID != nil {
		t.Errorf("root = %+v, %v; want B at the root", root.GetResult(), root.GetErr())
	}

	tests := []struct {
		name  string
		input Input
		want  error
	}{
		{"parent is a note", Input{IDs: []string{a}, ParentID: note.GetResult().ID}, ErrParentNotFound},
		{"unknown parent", Input{IDs: []string{a}, ParentID: unknown}, ErrParentNotFound},
		{"into itself", Input{IDs: []string{b}, ParentID: b}, ErrCycle},
		{"into its child", Input{IDs: []string{b}, ParentID: child}, ErrCycle},
		// Last kind of case: a DB error (RunInTx uses a savepoint here, but keep them at the end).
		{"not a uuid", Input{IDs: []string{"nope"}, ParentID: a}, ErrInvalidID},
	}
	for _, test := range tests {
		process := Process(ctx, tx, test.input)
		process.Run()
		if !errors.Is(process.GetErr(), test.want) {
			t.Errorf("%s: err = %v, want %v", test.name, process.GetErr(), test.want)
		}
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
