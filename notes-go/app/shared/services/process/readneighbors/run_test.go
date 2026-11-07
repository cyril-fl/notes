package readneighbors

import (
	"context"
	"errors"
	"fmt"
	"main.go/app/shared/services/testaccount"
	"slices"
	"testing"
	"time"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/deleteitems"
	"main.go/app/shared/services/process/readitem"
)

func TestNeighbors(t *testing.T) {
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

	// Unique per run so real data in the DB cannot collide.
	a, b, c := fmt.Sprintf("test-%d-a", time.Now().UnixNano()), fmt.Sprintf("test-%d-b", time.Now().UnixNano()), fmt.Sprintf("test-%d-c", time.Now().UnixNano())
	ids := map[string]string{}
	for _, note := range []struct {
		title string
		tags  []string
	}{
		{"Center", []string{a, b}},
		{"Two", []string{a, b, c}},
		{"One", []string{b}},
		{"None", []string{c}},
		{"Trashed", []string{a}},
	} {
		process := createnote.Process(ctx, tx, createnote.Input{Content: "# " + note.title, Tags: note.tags})
		process.Run()
		if process.HasErr() {
			t.Fatalf("note %s: %v", note.title, process.GetErr())
		}
		ids[note.title] = process.GetResult().ID
	}
	trash := deleteitems.Process(ctx, tx, domain.NOTE, []string{ids["Trashed"]})
	trash.Run()
	if trash.HasErr() {
		t.Fatalf("trash: %v", trash.GetErr())
	}

	// Two shares a and b, One shares b; None shares nothing, Trashed is in the trash.
	process := Process(ctx, tx, ids["Center"])
	process.Run()
	got := process.GetResult()
	if process.HasErr() || got.Note.Item.Title != "Center" || len(got.Neighbors) != 2 {
		t.Fatalf("result = %+v, %v; want Center with 2 neighbors", got, process.GetErr())
	}
	if two := got.Neighbors[0]; two.ID != ids["Two"] || two.Shared != 2 || !slices.Equal(two.Tags, []string{a, b}) || two.Path != "/Two" {
		t.Errorf("first neighbor = %+v, want Two sharing [a b]", two)
	}
	if one := got.Neighbors[1]; one.ID != ids["One"] || one.Shared != 1 {
		t.Errorf("second neighbor = %+v, want One sharing 1", one)
	}

	missing := Process(ctx, tx, "00000000-0000-0000-0000-000000000000")
	missing.Run()
	if !errors.Is(missing.GetErr(), readitem.ErrNotFound) {
		t.Errorf("err = %v, want %v", missing.GetErr(), readitem.ErrNotFound)
	}
}
