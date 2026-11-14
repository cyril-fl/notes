package countnotes

import (
	"context"
	"testing"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/deleteitems"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
)

func TestCountNotes(t *testing.T) {
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

	var trashedID string
	for _, title := range []string{"A", "B", "Trashed"} {
		note := createnote.Process(ctx, tx, createnote.Input{Content: "# " + title})
		note.Run()
		if note.HasErr() {
			t.Fatalf("note %s: %v", title, note.GetErr())
		}
		trashedID = note.GetResult().ID
	}
	trash := deleteitems.Process(ctx, tx, domain.NOTE, []string{trashedID})
	trash.Run()
	if trash.HasErr() {
		t.Fatalf("trash: %v", trash.GetErr())
	}

	process := Process(ctx, tx)
	process.Run()
	if stats := process.GetResult(); process.HasErr() || stats.Notes != 2 || stats.Size <= 0 {
		t.Errorf("stats = %+v, %v; want 2 notes and a size", stats, process.GetErr())
	}
}
