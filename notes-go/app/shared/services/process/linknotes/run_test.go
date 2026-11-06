// External package: the processes driving links (createnote, editnote) import linknotes.
package linknotes_test

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createfolder"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/editnote"
	"main.go/app/shared/services/process/readitem"
	"main.go/app/shared/services/process/readneighbors"
)

func TestLinks(t *testing.T) {
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

	// Unique per run so real notes cannot match the titles.
	u := fmt.Sprintf("test-%d", time.Now().UnixNano())
	folder := createfolder.Process(ctx, tx, createfolder.Input{Title: u + "-folder"})
	folder.Run()
	if folder.HasErr() {
		t.Fatalf("folder: %v", folder.GetErr())
	}

	// a names b before b exists, itself, nothing, and the ambiguous c by title then by path.
	a := create(t, ctx, tx, "", fmt.Sprintf("# %[1]s-a\n\n[[%[1]s-B]] [[%[1]s-b]] [[%[1]s-a]] [[%[1]s-none]] [[%[1]s-c]] [[/%[1]s-folder/%[1]s-c]]", u))
	b := create(t, ctx, tx, "", "# "+u+"-b")
	create(t, ctx, tx, "", "# "+u+"-c")
	inFolder := create(t, ctx, tx, folder.GetResult().ID, "# "+u+"-c")

	// b got its backlink when it was created; [[c]] matches two notes, its path only one.
	got := neighbors(t, ctx, tx, a)
	if !slices.Equal(ids(got.Links), sorted(b, inFolder)) || len(got.Backlinks) != 0 {
		t.Errorf("links of a = %+v, backlinks %+v; want b and the c in the folder", got.Links, got.Backlinks)
	}
	if got := neighbors(t, ctx, tx, b); !slices.Equal(ids(got.Backlinks), []string{a}) {
		t.Errorf("backlinks of b = %+v, want a", got.Backlinks)
	}

	// Renaming b rewrites the links of a, whatever their case.
	edit := editnote.Process(ctx, tx, editnote.Input{ID: b, Content: "# " + u + "-renamed"})
	edit.Run()
	if edit.HasErr() {
		t.Fatalf("rename: %v", edit.GetErr())
	}
	read := readitem.Process(ctx, tx, domain.NOTE, a)
	read.Run()
	want := fmt.Sprintf("[[%[1]s-renamed]] [[%[1]s-renamed]] [[%[1]s-a]]", u)
	if read.HasErr() || !strings.Contains(read.GetResult().Content, want) {
		t.Errorf("content of a = %q, %v; want it to hold %q", read.GetResult().Content, read.GetErr(), want)
	}
	if got := neighbors(t, ctx, tx, a); !slices.Contains(ids(got.Links), b) {
		t.Errorf("links of a after the rename = %+v, want b still", got.Links)
	}
}

// Helpers

func create(t *testing.T, ctx context.Context, tx bun.Tx, parentID, content string) string {
	t.Helper()
	note := createnote.Process(ctx, tx, createnote.Input{Content: content, ParentID: parentID})
	note.Run()
	if note.HasErr() {
		t.Fatalf("create %q: %v", content, note.GetErr())
	}
	return note.GetResult().ID
}

func neighbors(t *testing.T, ctx context.Context, tx bun.Tx, id string) readneighbors.Result {
	t.Helper()
	read := readneighbors.Process(ctx, tx, id)
	read.Run()
	if read.HasErr() {
		t.Fatalf("neighbors: %v", read.GetErr())
	}
	return read.GetResult()
}

func ids(notes []readneighbors.Linked) []string {
	var result []string
	for _, note := range notes {
		result = append(result, note.ID)
	}
	slices.Sort(result)
	return result
}

func sorted(values ...string) []string {
	slices.Sort(values)
	return values
}
