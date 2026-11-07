package importnotes

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createfolder"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/exportnotes"
)

func TestImportNotes(t *testing.T) {
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

	// Unique per run: the dev database may hold anything at the root.
	create := createfolder.Process(ctx, tx, createfolder.Input{Title: fmt.Sprintf("import-%d", time.Now().UnixNano())})
	create.Run()
	if create.HasErr() {
		t.Fatal(create.GetErr())
	}
	parent := create.GetResult().ID

	dir := fstest.MapFS{
		"Projets/Idées.md":    {Data: []byte("# Idées\n\nvoir [[Plan unique 7f3]] #go")},
		"Projets/Plan.md":     {Data: []byte("# Plan unique 7f3\n\ntexte")},
		"Projets/image.png":   {Data: []byte("png")},
		"Projets/vide.md":     {Data: []byte("  \n")},
		".obsidian/config.md": {Data: []byte("# Config")},
		"Vide/.gitkeep":       {Data: []byte("")},
		"Racine.md":           {Data: []byte("du texte")},
		"du lait.md":          {Data: []byte("du lait")},
	}
	run := Process(ctx, tx, Input{Dir: dir, ParentID: parent})
	run.Run()
	if run.HasErr() {
		t.Fatalf("import: %v", run.GetErr())
	}
	if got, want := run.GetResult(), (Result{Folders: 2, Notes: 4}); got != want {
		t.Errorf("result = %+v, want %+v", got, want)
	}

	// The export of what came in gives the files back, the link resolved on the way and written back as typed.
	export := exportnotes.Process(ctx, tx, parent)
	export.Run()
	if export.HasErr() {
		t.Fatal(export.GetErr())
	}
	var got []string
	for _, file := range export.GetResult() {
		got = append(got, file.Path+"="+file.Content)
	}
	want := []string{
		"Projets=",
		"Projets/Idées.md=# Idées\n\nvoir [[Plan unique 7f3]] #go",
		"Projets/Plan unique 7f3.md=# Plan unique 7f3\n\ntexte",
		"Vide=",
		"du lait.md=du lait",
		"Racine.md=# Racine\n\ndu texte",
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("export = %q, want %q", got, want)
	}

	linked, err := tx.NewSelect().Model((*domain.NoteLink)(nil)).
		Join("JOIN items AS t ON t.id = nl.target_id").Where("t.title = ? AND t.parent_id IS NOT NULL", "Plan unique 7f3").
		Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if linked != 1 {
		t.Errorf("links to the imported Plan = %d, want 1", linked)
	}

	// A folder name taken stops it all: nothing of the second import stays.
	again := Process(ctx, tx, Input{Dir: dir, ParentID: parent})
	again.Run()
	if !errors.Is(again.GetErr(), domain.ErrFolderExists) {
		t.Errorf("second import: err = %v, want %v", again.GetErr(), domain.ErrFolderExists)
	}
}
