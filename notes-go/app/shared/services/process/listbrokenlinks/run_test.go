package listbrokenlinks

import (
	"context"
	"slices"
	"testing"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/deleteitems"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
)

func TestBrokenLinks(t *testing.T) {
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

	ids := map[string]string{}
	for _, note := range []struct{ key, content string }{
		{"Twin1", "# Twin"},
		{"Twin2", "# twin"},
		{"Single", "# Single\n\n## Here"},
		{"Gone", "# Gone"},
		// Single and Gone resolve on creation, the rest waits.
		{"Source", "# Source\n\n[[Twin#h]] [[Nowhere]] [[Single]] [[Single#Here]] [[Single#Away]] [[Gone]] [[#Top]] [[#Gone (old)]] [[Source]] `[[Code]]`\n\n## Top"},
	} {
		process := createnote.Process(ctx, tx, createnote.Input{Content: note.content})
		process.Run()
		if process.HasErr() {
			t.Fatalf("note %s: %v", note.key, process.GetErr())
		}
		ids[note.key] = process.GetResult().ID
	}
	trash := deleteitems.Process(ctx, tx, domain.NOTE, []string{ids["Gone"]})
	trash.Run()
	if trash.HasErr() {
		t.Fatalf("trash: %v", trash.GetErr())
	}

	process := Process(ctx, tx)
	process.Run()
	if process.HasErr() {
		t.Fatalf("list: %v", process.GetErr())
	}

	var got []string
	for _, link := range process.GetResult() {
		if link.NoteID != ids["Source"] || link.NoteTitle != "Source" || link.NotePath != "/Source" {
			t.Errorf("link %+v: wrong note", link)
		}
		got = append(got, link.Kind+":"+link.Target)
	}
	// [[Source]] names only the note itself: it never resolves.
	want := []string{"ambiguous:Twin#h", "unresolved:Nowhere", "unresolved:#Gone (old)", "unresolved:Source", "unresolved:Single#Away", "missing:Gone"}
	if !slices.Equal(got, want) {
		t.Errorf("links = %q, want %q", got, want)
	}
}
