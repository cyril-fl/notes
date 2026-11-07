package syncentities

import (
	"context"
	"testing"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/readitem"
)

func TestSyncEntities(t *testing.T) {
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

	// As the editor saved them before the fix: the title follows the text, entities included.
	seeds := []struct{ content, wantContent, wantTitle string }{
		{"# a &gt;&gt; b\n\nc &amp; d", "# a >> b\n\nc & d", "a >> b"},
		{"# Code\n\n`&lt;`", "# Code\n\n`&lt;`", "Code"},
	}
	ids := make([]string, len(seeds))
	for i, seed := range seeds {
		note := createnote.Process(ctx, tx, createnote.Input{Content: seed.content})
		note.Run()
		if note.HasErr() {
			t.Fatalf("seed %d: %v", i, note.GetErr())
		}
		ids[i] = note.GetResult().ID
	}

	first := Process(ctx, tx)
	first.Run()
	if first.HasErr() {
		t.Fatalf("first run: %v", first.GetErr())
	}
	for i, seed := range seeds {
		read := readitem.Process(ctx, tx, domain.NOTE, ids[i])
		read.Run()
		if got := read.GetResult(); got.Content != seed.wantContent || got.Item.Title != seed.wantTitle {
			t.Errorf("note %d = %q / %q, want %q / %q", i, got.Content, got.Item.Title, seed.wantContent, seed.wantTitle)
		}
	}

	second := Process(ctx, tx)
	second.Run()
	if second.HasErr() || second.GetResult() != 0 {
		t.Errorf("second run changed %d note(s), err %v; want 0", second.GetResult(), second.GetErr())
	}
}
