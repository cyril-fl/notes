package synctags

import (
	"context"
	"slices"
	"strings"
	"testing"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/createtag"
	"main.go/app/shared/services/process/readitem"
)

func TestSyncBothWays(t *testing.T) {
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

	seed := createnote.Process(ctx, tx, createnote.Input{Content: "texte"})
	seed.Run()
	if seed.HasErr() {
		t.Fatalf("seed: %v", seed.GetErr())
	}
	id := seed.GetResult().ID

	// A note as it was before: one link with no #tag in the text, one #tag with no link.
	tag := createtag.Process(ctx, tx, []string{"synclinked"})
	tag.Run()
	if tag.HasErr() {
		t.Fatalf("tag: %v", tag.GetErr())
	}
	if _, err := tx.NewInsert().Model(&domain.NoteTag{TagID: tag.GetResult()[0].ID, NoteID: id}).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.NewUpdate().Model((*domain.MetadataNote)(nil)).Set("content = ?", "texte #synctext").Where("item_id = ?", id).Exec(ctx); err != nil {
		t.Fatal(err)
	}

	first := Process(ctx, tx)
	first.Run()
	if first.HasErr() {
		t.Fatalf("first run: %v", first.GetErr())
	}
	if first.GetResult() == 0 {
		t.Error("first run changed nothing")
	}

	read := readitem.Process(ctx, tx, domain.NOTE, id)
	read.Run()
	got := read.GetResult()
	if read.HasErr() || !strings.Contains(got.Content, "#synctext") || !strings.Contains(got.Content, "#synclinked") {
		t.Errorf("content = %q (err %v)", got.Content, read.GetErr())
	}
	if !slices.Equal(got.Tags, []string{"synclinked", "synctext"}) {
		t.Errorf("tags = %v", got.Tags)
	}

	second := Process(ctx, tx)
	second.Run()
	if second.HasErr() || second.GetResult() != 0 {
		t.Errorf("second run changed %d notes (err %v)", second.GetResult(), second.GetErr())
	}
}
