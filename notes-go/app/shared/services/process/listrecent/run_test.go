package listrecent

import (
	"context"
	"fmt"
	"main.go/app/shared/services/testaccount"
	"testing"
	"time"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/deleteitems"
)

func TestListRecent(t *testing.T) {
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

	// Dated in the future so real notes rank after; current_timestamp would be the same for the whole transaction.
	future := time.Now().AddDate(100, 0, 0)
	note := func(title string, age int) string {
		process := createnote.Process(ctx, tx, createnote.Input{Content: "# " + title})
		process.Run()
		if process.HasErr() {
			t.Fatalf("note %s: %v", title, process.GetErr())
		}
		id := process.GetResult().ID
		if _, err := tx.NewRaw("UPDATE items SET created_at = ? WHERE id = ?", future.Add(-time.Duration(age)*time.Hour), id).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		return id
	}

	for i := range limit + 1 {
		note(fmt.Sprintf("note-%02d", i), i)
	}

	// Created long ago but just edited: updated_at wins over created_at.
	edited := note("edited", 500)
	if _, err := tx.NewRaw("UPDATE items SET updated_at = ? WHERE id = ?", future.Add(time.Hour), edited).Exec(ctx); err != nil {
		t.Fatal(err)
	}

	// The newest of all, but trashed: it must not show.
	trashed := note("trashed", -10)
	trash := deleteitems.Process(ctx, tx, domain.NOTE, []string{trashed})
	trash.Run()
	if trash.HasErr() {
		t.Fatalf("trash: %v", trash.GetErr())
	}

	process := Process(ctx, tx)
	process.Run()
	if process.HasErr() {
		t.Fatal(process.GetErr())
	}

	result := process.GetResult()
	if len(result) != limit {
		t.Fatalf("len = %d, want %d", len(result), limit)
	}
	if result[0].Title != "edited" || result[1].Title != "note-00" || result[limit-1].Title != fmt.Sprintf("note-%02d", limit-2) {
		t.Errorf("order = %s, %s, ... %s", result[0].Title, result[1].Title, result[limit-1].Title)
	}
	if result[1].Path != "/note-00" {
		t.Errorf("path = %q, want /note-00", result[1].Path)
	}
}
