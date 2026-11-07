package synctitles

import (
	"context"
	"main.go/app/shared/services/testaccount"
	"testing"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/readitem"
)

func TestSyncTitles(t *testing.T) {
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

	// Notes as they were before: a title typed apart from the text, or none.
	seeds := []struct{ content, title, wantContent, wantTitle string }{
		{"du pain", "Courses", "# Courses\n\ndu pain", "Courses"},
		{"# Plan\n\ntexte", "Plan", "# Plan\n\ntexte", "Plan"},
		{"du lait", "", "du lait", "du lait"},
	}
	ids := make([]string, len(seeds))
	for i, seed := range seeds {
		note := createnote.Process(ctx, tx, createnote.Input{Content: seed.content})
		note.Run()
		if note.HasErr() {
			t.Fatalf("seed %d: %v", i, note.GetErr())
		}
		ids[i] = note.GetResult().ID
		if _, err := tx.NewUpdate().Model((*domain.Item)(nil)).Set("title = NULLIF(?, '')", seed.title).Where("id = ?", ids[i]).Exec(ctx); err != nil {
			t.Fatal(err)
		}
	}

	first := Process(ctx, tx)
	first.Run()
	if first.HasErr() {
		t.Fatalf("first run: %v", first.GetErr())
	}
	for i, seed := range seeds {
		read := readitem.Process(ctx, tx, domain.NOTE, ids[i])
		read.Run()
		got := read.GetResult()
		if read.HasErr() || got.Content != seed.wantContent || got.Item.Title != seed.wantTitle {
			t.Errorf("note %d = %q, %q, %v; want %q, %q", i, got.Content, got.Item.Title, read.GetErr(), seed.wantContent, seed.wantTitle)
		}
	}

	second := Process(ctx, tx)
	second.Run()
	if second.HasErr() || second.GetResult() != 0 {
		t.Errorf("second run changed %d, %v; want 0", second.GetResult(), second.GetErr())
	}
}
