package editnote

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createfolder"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/readitem"
)

// No DB needed: makeNote fails first and saveNote is guarded, so a nil db is never touched.
func TestNoValue(t *testing.T) {
	tests := []struct {
		input Input
		want  error
	}{
		{Input{ID: " ", Content: "C"}, ErrNoValue},
		{Input{ID: "id", Content: " "}, ErrNoContent},
	}

	for _, test := range tests {
		process := Process(context.Background(), nil, test.input)
		process.Run()
		if !errors.Is(process.GetErr(), test.want) {
			t.Errorf("%+v: err = %v, want %v", test.input, process.GetErr(), test.want)
		}
	}
}

func TestEditNote(t *testing.T) {
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

	// Unique per run so real data in the DB cannot collide.
	tag := fmt.Sprintf("test-%d", time.Now().UnixNano())
	seed := createnote.Process(ctx, tx, createnote.Input{Title: "Old", Content: "old", Tags: []string{tag, tag + "-old"}})
	seed.Run()
	if seed.HasErr() {
		t.Fatalf("seed: %v", seed.GetErr())
	}
	id := seed.GetResult().ID

	edit := Process(ctx, tx, Input{ID: id, Title: " ", Content: " new ", Tags: []string{tag, tag + "-new"}})
	edit.Run()
	if edit.HasErr() {
		t.Fatalf("edit: %v", edit.GetErr())
	}
	if item := edit.GetResult(); item.ID != id || item.Title != "" || item.UpdatedAt.IsZero() {
		t.Errorf("result = %+v, want same id, untitled, updated_at set", item)
	}

	read := readitem.Process(ctx, tx, domain.NOTE, id)
	read.Run()
	got := read.GetResult()
	want := "new\n\n#" + tag + " #" + tag + "-new"
	if read.HasErr() || got.Content != want || !slices.Equal(got.Tags, []string{tag, tag + "-new"}) {
		t.Errorf("read = %+v, %v; want new content and tags replaced", got, read.GetErr())
	}

	// The tag this edit unlinked had no other note.
	left, err := tx.NewSelect().Model((*domain.Tag)(nil)).Where("value = ?", tag+"-old").Count(ctx)
	if err != nil || left != 0 {
		t.Errorf("unused tag left = %d, %v; want 0", left, err)
	}

	// Removing the tags from the text of a note holding nothing else leaves it empty: refused.
	bare := Process(ctx, tx, Input{ID: id, Content: "#" + tag})
	bare.Run()
	if !errors.Is(bare.GetErr(), ErrNoContent) {
		t.Errorf("empty once synced: err = %v, want %v", bare.GetErr(), ErrNoContent)
	}

	folder := createfolder.Process(ctx, tx, createfolder.Input{Title: "Folder"})
	folder.Run()
	if folder.HasErr() {
		t.Fatalf("folder: %v", folder.GetErr())
	}

	// Last: a folder id, an unknown id and a non-uuid id are all not found.
	for _, id := range []string{folder.GetResult().ID, "00000000-0000-0000-0000-000000000000", "nope"} {
		missing := Process(ctx, tx, Input{ID: id, Content: "C"})
		missing.Run()
		if !errors.Is(missing.GetErr(), ErrNotFound) {
			t.Errorf("%s: err = %v, want %v", id, missing.GetErr(), ErrNotFound)
		}
	}
}
