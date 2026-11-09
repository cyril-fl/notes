package exportnotes

import (
	"context"
	"errors"
	"fmt"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
	"slices"
	"testing"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createfolder"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
)

func TestExportNotes(t *testing.T) {
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

	folder := func(title, parent string) string {
		create := createfolder.Process(ctx, tx, createfolder.Input{Title: title, ParentID: parent})
		create.Run()
		if create.HasErr() {
			t.Fatalf("folder %q: %v", title, create.GetErr())
		}
		return create.GetResult().ID
	}
	note := func(content, parent string) string {
		create := createnote.Process(ctx, tx, createnote.Input{Content: content, ParentID: parent})
		create.Run()
		if create.HasErr() {
			t.Fatalf("note %q: %v", content, create.GetErr())
		}
		return create.GetResult().ID
	}

	// Unique per run: the dev database may hold anything at the root.
	root := folder(fmt.Sprintf("export-%d", time.Now().UnixNano()), "")
	sub := folder("Sous/dossier", root)
	note("# Courses\n\ndu pain", root)
	milk := note("# courses\n\ndu lait", root)
	note("voir [[Courses]]", sub)

	export := Process(ctx, tx, root)
	export.Run()
	if export.HasErr() {
		t.Fatalf("export: %v", export.GetErr())
	}

	want := []File{
		{Path: "Sous-dossier", Dir: true},
		// Ambiguous, the link stays waiting: exported as typed.
		{Path: "Sous-dossier/voir Courses.md", Content: "voir [[Courses]]"},
		{Path: "Courses.md", Content: "# Courses\n\ndu pain"},
		{Path: "courses (2).md", Content: "# courses\n\ndu lait"},
	}
	if got := export.GetResult(); !slices.Equal(got, want) {
		t.Errorf("files = %+v, want %+v", got, want)
	}

	// A selection: the folders with what they hold, at the top of the export.
	selection := Process(ctx, tx, "", sub, milk)
	selection.Run()
	if selection.HasErr() {
		t.Fatalf("export the selection: %v", selection.GetErr())
	}
	want = []File{
		{Path: "Sous-dossier", Dir: true},
		{Path: "Sous-dossier/voir Courses.md", Content: "voir [[Courses]]"},
		{Path: "courses.md", Content: "# courses\n\ndu lait"},
	}
	if got := selection.GetResult(); !slices.Equal(got, want) {
		t.Errorf("selection = %+v, want %+v", got, want)
	}
	unknown := Process(ctx, tx, "", milk, "00000000-0000-0000-0000-000000000000")
	unknown.Run()
	if !errors.Is(unknown.GetErr(), ErrNotFound) {
		t.Errorf("export of a missing item: err = %v, want ErrNotFound", unknown.GetErr())
	}

	missing := Process(ctx, tx, "00000000-0000-0000-0000-000000000000")
	missing.Run()
	if !missing.HasErr() {
		t.Error("export of a missing folder: want an error")
	}
}

func TestFileName(t *testing.T) {
	cases := map[string]string{"Idées": "Idées", "a/b": "a-b", ".env": "env", "  ": "untitled", "": "untitled"}
	for title, want := range cases {
		if got := fileName(title); got != want {
			t.Errorf("fileName(%q) = %q, want %q", title, got, want)
		}
	}
}
