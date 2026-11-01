package readtag

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createfolder"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/deleteitems"
)

// No DB needed: makeValue fails first and the loads are guarded, so a nil db is never touched.
func TestNoValue(t *testing.T) {
	process := Process(context.Background(), nil, " ")
	process.Run()
	if !errors.Is(process.GetErr(), ErrNoValue) {
		t.Fatalf("err = %v, want %v", process.GetErr(), ErrNoValue)
	}
}

func TestReadTag(t *testing.T) {
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

	// Unique per run so real data in the DB cannot collide.
	tag := fmt.Sprintf("test-%d", time.Now().UnixNano())
	folder := createfolder.Process(ctx, tx, createfolder.Input{Title: "F"})
	folder.Run()
	if folder.HasErr() {
		t.Fatalf("folder: %v", folder.GetErr())
	}
	var ids []string
	for _, input := range []createnote.Input{
		{Title: "Root", Content: "C", Tags: []string{tag}},
		{Title: "Inside", Content: "C", Tags: []string{tag}, ParentID: folder.GetResult().ID},
		{Title: "Trashed", Content: "C", Tags: []string{tag}},
	} {
		note := createnote.Process(ctx, tx, input)
		note.Run()
		if note.HasErr() {
			t.Fatalf("note %s: %v", input.Title, note.GetErr())
		}
		ids = append(ids, note.GetResult().ID)
	}
	trash := deleteitems.Process(ctx, tx, domain.NOTE, ids[2:])
	trash.Run()
	if trash.HasErr() {
		t.Fatalf("trash: %v", trash.GetErr())
	}

	read := Process(ctx, tx, strings.ToUpper(tag))
	read.Run()
	got := read.GetResult()
	if read.HasErr() || got.Tag.Value != tag || len(got.Notes) != 2 || got.Notes[0].Path != "/F/Inside" || got.Notes[1].Path != "/Root" {
		t.Errorf("result = %+v, %v; want %s with /F/Inside, /Root", got, read.GetErr(), tag)
	}

	missing := Process(ctx, tx, tag+"-missing")
	missing.Run()
	if !errors.Is(missing.GetErr(), ErrNotFound) {
		t.Errorf("err = %v, want %v", missing.GetErr(), ErrNotFound)
	}
}
