package readsavedsearch

import (
	"context"
	"errors"
	"testing"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/edituser"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
)

func TestReadSavedSearch(t *testing.T) {
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

	note := createnote.Process(ctx, tx, createnote.Input{Content: "# Match me"})
	note.Run()
	if note.HasErr() {
		t.Fatalf("note: %v", note.GetErr())
	}

	searches := []domain.SavedSearch{{ID: "s1", Name: "Everything"}}
	save := edituser.Process(ctx, tx, edituser.Input{SavedSearches: &searches})
	save.Run()
	if save.HasErr() {
		t.Fatalf("save: %v", save.GetErr())
	}

	found := Process(ctx, tx, "s1")
	found.Run()
	if found.HasErr() || len(found.GetResult()) != 1 || found.GetResult()[0].Title != "Match me" {
		t.Fatalf("found = %+v, %v; want the note", found.GetResult(), found.GetErr())
	}

	missing := Process(ctx, tx, "absent")
	missing.Run()
	if !errors.Is(missing.GetErr(), ErrNotFound) {
		t.Fatalf("missing: err = %v, want %v", missing.GetErr(), ErrNotFound)
	}
}
