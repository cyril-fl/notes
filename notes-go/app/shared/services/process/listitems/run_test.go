package listitems

import (
	"context"
	"errors"
	"main.go/app/shared/services/testaccount"
	"slices"
	"testing"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createfolder"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/deleteitems"
)

func TestListItems(t *testing.T) {
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

	parent := createfolder.Process(ctx, tx, createfolder.Input{Title: "Parent"})
	parent.Run()
	if parent.HasErr() {
		t.Fatalf("parent: %v", parent.GetErr())
	}
	id := parent.GetResult().ID

	note := createnote.Process(ctx, tx, createnote.Input{Content: "# A note", ParentID: id})
	note.Run()
	sub := createfolder.Process(ctx, tx, createfolder.Input{Title: "Z folder", ParentID: id})
	sub.Run()
	trashed := createfolder.Process(ctx, tx, createfolder.Input{Title: "Trashed", ParentID: id})
	trashed.Run()
	trash := deleteitems.Process(ctx, tx, domain.FOLDER, []string{trashed.GetResult().ID})
	trash.Run()
	if note.HasErr() || sub.HasErr() || trashed.HasErr() || trash.HasErr() {
		t.Fatalf("seed: %v %v %v %v", note.GetErr(), sub.GetErr(), trashed.GetErr(), trash.GetErr())
	}

	// Folders first even when a note sorts before them by title; the trashed folder is gone.
	children := Process(ctx, tx, id)
	children.Run()
	var got []string
	for _, item := range children.GetResult() {
		got = append(got, string(item.Title))
	}
	if children.HasErr() || !slices.Equal(got, []string{"Z folder", "A note"}) {
		t.Errorf("children = %v, %v; want [Z folder, A note]", got, children.GetErr())
	}

	root := Process(ctx, tx, "")
	root.Run()
	found := slices.ContainsFunc(root.GetResult(), func(item domain.Item) bool { return item.ID == id })
	if root.HasErr() || !found {
		t.Errorf("root = %v; want Parent in it", root.GetErr())
	}

	// Last: a DB error.
	invalid := Process(ctx, tx, "nope")
	invalid.Run()
	if !errors.Is(invalid.GetErr(), ErrInvalidID) {
		t.Errorf("err = %v, want %v", invalid.GetErr(), ErrInvalidID)
	}
}
