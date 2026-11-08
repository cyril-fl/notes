package readtag

import (
	"context"
	"errors"
	"fmt"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
	"strings"
	"testing"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createfolder"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/deleteitems"
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
	ctx, tx = testaccount.New(t, ctx, tx)

	// Unique per run so real data in the DB cannot collide.
	tag := fmt.Sprintf("test-%d", time.Now().UnixNano())
	folder := createfolder.Process(ctx, tx, createfolder.Input{Title: "F-" + tag})
	folder.Run()
	if folder.HasErr() {
		t.Fatalf("folder: %v", folder.GetErr())
	}
	var ids []string
	for _, input := range []createnote.Input{
		{Content: "# Root", Tags: []string{tag}},
		{Content: "# Inside", Tags: []string{tag}, ParentID: folder.GetResult().ID},
		{Content: "# Trashed", Tags: []string{tag}},
	} {
		note := createnote.Process(ctx, tx, input)
		note.Run()
		if note.HasErr() {
			t.Fatalf("note %s: %v", input.Content, note.GetErr())
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
	if read.HasErr() || string(got.Tag.Value) != tag || len(got.Notes) != 2 || got.Notes[0].Path != "/F-"+tag+"/Inside" || got.Notes[1].Path != "/Root" {
		t.Errorf("result = %+v, %v; want %s with /F/Inside, /Root", got, read.GetErr(), tag)
	}

	missing := Process(ctx, tx, tag+"-missing")
	missing.Run()
	if !errors.Is(missing.GetErr(), ErrNotFound) {
		t.Errorf("err = %v, want %v", missing.GetErr(), ErrNotFound)
	}
}
