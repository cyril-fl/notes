package keepconflict

import (
	"context"
	"errors"
	"testing"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
)

func TestKeepConflict(t *testing.T) {
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

	seed := createnote.Process(ctx, tx, createnote.Input{Content: "theirs"})
	seed.Run()
	if seed.HasErr() {
		t.Fatalf("seed: %v", seed.GetErr())
	}
	id := seed.GetResult().ID
	keep := func(content string) *run {
		process := Process(ctx, tx, id, content)
		process.Run()
		return process
	}

	// The refused text is kept, marked conflict, the note left as it is.
	if got := keep("mine"); got.HasErr() || got.GetResult().ID == "" || !got.GetResult().Conflict || got.GetResult().Content != "mine" {
		t.Errorf("keep: %+v, %v", got.GetResult(), got.GetErr())
	}
	var saved domain.NoteVersion
	if err := tx.NewSelect().Model(&saved).Where("item_id = ?", id).Scan(ctx); err != nil || saved.Content != "mine" || !saved.Conflict {
		t.Errorf("saved = %+v, %v", saved, err)
	}

	// The note's own text: nothing to keep.
	if got := keep("theirs"); got.HasErr() || got.GetResult().ID != "" {
		t.Errorf("same text: %+v, %v", got.GetResult(), got.GetErr())
	}

	// Another account's note, as a missing one.
	unknown := Process(ctx, tx, "00000000-0000-0000-0000-000000000000", "mine")
	unknown.Run()
	if got := unknown; !errors.Is(got.GetErr(), readitem.ErrNotFound) {
		t.Errorf("unknown note: %v", got.GetErr())
	}
}
