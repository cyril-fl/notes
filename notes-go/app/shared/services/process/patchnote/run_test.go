package patchnote

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/hashtag"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
)

func TestPatchNote(t *testing.T) {
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

	seed := createnote.Process(ctx, tx, createnote.Input{Content: "old", Tags: []string{"a"}})
	seed.Run()
	if seed.HasErr() {
		t.Fatalf("seed: %v", seed.GetErr())
	}
	id := seed.GetResult().ID
	patch := func(input Input) *run {
		input.ID = id
		process := Process(ctx, tx, input)
		process.Run()
		return process
	}
	yes, content, bad := true, "new #b", []string{"a b"}

	// Pinning alone leaves the text and updated_at as they are.
	if got := patch(Input{Pinned: &yes}); got.HasErr() || !got.GetResult().Pinned || got.GetResult().Content != "old\n\n#a" || got.GetResult().Item.UpdatedAt != seed.GetResult().UpdatedAt {
		t.Errorf("pin only: %+v, %v", got.GetResult(), got.GetErr())
	}

	// The content alone: its #tags are the tags.
	if got := patch(Input{Content: &content}); got.HasErr() || !slices.Equal(got.GetResult().Tags, []string{"b"}) {
		t.Errorf("content only: %+v, %v", got.GetResult(), got.GetErr())
	}

	// A failed edit undoes the pin of the same call (#39).
	no := false
	if got := patch(Input{Pinned: &no, Tags: &bad}); !errors.Is(got.GetErr(), hashtag.ErrInvalid) {
		t.Errorf("bad tags: err = %v, want %v", got.GetErr(), hashtag.ErrInvalid)
	}
	read := readitem.Process(ctx, tx, domain.NOTE, id)
	read.Run()
	if !read.GetResult().Pinned {
		t.Errorf("a failed patch unpinned the note: %+v", read.GetResult())
	}

	// A locked note refuses its content, unless unlocked in the same call (#60).
	if got := patch(Input{Readonly: &yes}); got.HasErr() || !got.GetResult().Readonly || got.GetResult().Content != "new #b" {
		t.Errorf("lock only: %+v, %v", got.GetResult(), got.GetErr())
	}
	if got := patch(Input{Content: &content, Tags: &[]string{"c"}}); !errors.Is(got.GetErr(), ErrLocked) {
		t.Errorf("edit locked: err = %v, want %v", got.GetErr(), ErrLocked)
	}
	if got := patch(Input{Readonly: &no, Tags: &[]string{"c"}}); got.HasErr() || got.GetResult().Readonly || !slices.Equal(got.GetResult().Tags, []string{"c"}) {
		t.Errorf("unlock and edit: %+v, %v", got.GetResult(), got.GetErr())
	}
}
