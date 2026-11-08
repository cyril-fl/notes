package patchnote

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

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

	// The text replaced is kept, once every 5 minutes: the other edits of this transaction add none (#61).
	versions := func() (got []domain.NoteVersion) {
		if err := tx.NewSelect().Model(&got).Where("item_id = ?", id).OrderExpr("created_at DESC").Scan(ctx); err != nil {
			t.Fatal(err)
		}
		return got
	}
	if got := versions(); len(got) != 1 || got[0].Content != "old\n\n#a" {
		t.Errorf("versions after an edit: %+v", got)
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

	// A text written over an older version is refused, over the current one saved (#62).
	// current_timestamp is the start of the transaction, the same for every write here: an older version is made up.
	older := read.GetResult().Item.UpdatedAt.Add(-time.Second)
	if got := patch(Input{Content: &content, UpdatedAt: &older}); !errors.Is(got.GetErr(), ErrStale) {
		t.Errorf("stale: err = %v, want %v", got.GetErr(), ErrStale)
	}
	current := read.GetResult().Item.UpdatedAt
	if got := patch(Input{Content: &content, UpdatedAt: &current}); got.HasErr() {
		t.Errorf("current: %v", got.GetErr())
	}

	if got := versions(); len(got) != 1 {
		t.Errorf("versions within 5 minutes: %d, want 1", len(got))
	}

	// A restored version keeps the text it replaces, even within 5 minutes (#94).
	restored := "restored"
	if got := patch(Input{Content: &restored, Restore: true}); got.HasErr() {
		t.Fatalf("restore: %v", got.GetErr())
	}
	// One transaction, one current_timestamp: the versions tie on created_at, look for the text among them.
	if got := versions(); len(got) != 2 || !slices.ContainsFunc(got, func(v domain.NoteVersion) bool { return string(v.Content) == content }) {
		t.Errorf("versions after a restore: %+v, want 2 with %q", got, content)
	}
	if got := patch(Input{Content: &content}); got.HasErr() {
		t.Fatalf("back to content: %v", got.GetErr())
	}

	// Past 5 minutes the text is kept again, and only the versionsKept last stay.
	if _, err := tx.NewRaw("UPDATE note_versions SET created_at = created_at - interval '1 day' WHERE item_id = ?", id).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.NewRaw("INSERT INTO note_versions (item_id, content, created_at) SELECT item_id, content, created_at - g * interval '1 hour' FROM note_versions, generate_series(1, ?) AS g WHERE item_id = ?", versionsKept, id).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	again := "again"
	if got := patch(Input{Content: &again}); got.HasErr() {
		t.Fatalf("edit past 5 minutes: %v", got.GetErr())
	}
	if got := versions(); len(got) != versionsKept || string(got[0].Content) != content {
		t.Errorf("versions past 5 minutes: %d, newest %q", len(got), got[0].Content)
	}

	// A note never edited: its version is created_at.
	fresh := createnote.Process(ctx, tx, createnote.Input{Content: "fresh"})
	fresh.Run()
	created := fresh.GetResult().CreatedAt
	edit := Process(ctx, tx, Input{ID: fresh.GetResult().ID, Content: &content, UpdatedAt: &created})
	edit.Run()
	if edit.HasErr() {
		t.Errorf("never edited: %v", edit.GetErr())
	}
}
