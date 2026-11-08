// External package: the processes driving links (createnote, editnote) import linknotes.
package linknotes_test

import (
	"context"
	"fmt"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createfolder"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/editnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readneighbors"
	"github.com/uptrace/bun"
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
	ctx, tx = testaccount.New(t, ctx, tx)

	// Unique per run so real notes cannot match the titles.
	u := fmt.Sprintf("test-%d", time.Now().UnixNano())
	folder := createfolder.Process(ctx, tx, createfolder.Input{Title: u + "-folder"})
	folder.Run()
	if folder.HasErr() {
		t.Fatalf("folder: %v", folder.GetErr())
	}

	// a names b before b exists, itself, nothing, then c before a second c, by title and by path.
	a := create(t, ctx, tx, "", fmt.Sprintf("# %[1]s-a\n\n[[%[1]s-B]] [[%[1]s-b]] [[%[1]s-a]] [[%[1]s-none]] [[%[1]s-c]] [[/%[1]s-folder/%[1]s-c]]", u))
	b := create(t, ctx, tx, "", "# "+u+"-b")
	c := create(t, ctx, tx, "", "# "+u+"-c")
	inFolder := create(t, ctx, tx, folder.GetResult().ID, "# "+u+"-c")

	// Each [[link]] became a link to the id of the note it named when it was resolved: the second c
	// does not take the first one's link, its path reaches it.
	want := fmt.Sprintf("[%[1]s-b](note:%[2]s) [%[1]s-b](note:%[2]s) [[%[1]s-a]] [[%[1]s-none]] [%[1]s-c](note:%[3]s) [%[1]s-c](note:%[4]s)", u, b, c, inFolder)
	if got := content(t, ctx, tx, a); !strings.Contains(got, want) {
		t.Errorf("content of a = %q, want it to hold %q", got, want)
	}
	got := neighbors(t, ctx, tx, a)
	if !slices.Equal(ids(got.Links), sorted(b, c, inFolder)) || len(got.Backlinks) != 0 {
		t.Errorf("links of a = %+v, backlinks %+v; want b and both c", got.Links, got.Backlinks)
	}
	if got := neighbors(t, ctx, tx, b); !slices.Equal(ids(got.Backlinks), []string{a}) {
		t.Errorf("backlinks of b = %+v, want a", got.Backlinks)
	}

	// A title shared from the start resolves nothing: the link waits, with no row.
	d := create(t, ctx, tx, "", fmt.Sprintf("# %[1]s-d\n\n[[%[1]s-c]]", u))
	if got := content(t, ctx, tx, d); !strings.Contains(got, "[["+u+"-c]]") || len(neighbors(t, ctx, tx, d).Links) != 0 {
		t.Errorf("content of d = %q, want [[c]] unresolved", got)
	}

	// Renaming b keeps the links of a on it and writes its new title in them.
	edit := editnote.Process(ctx, tx, editnote.Input{ID: b, Content: "# " + u + "-renamed"})
	edit.Run()
	if edit.HasErr() {
		t.Fatalf("rename: %v", edit.GetErr())
	}
	want = fmt.Sprintf("[%[1]s-renamed](note:%[2]s) [%[1]s-renamed](note:%[2]s)", u, b)
	if got := content(t, ctx, tx, a); !strings.Contains(got, want) {
		t.Errorf("content of a = %q, want it to hold %q", got, want)
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

func content(t *testing.T, ctx context.Context, tx bun.Tx, id string) string {
	t.Helper()
	read := readitem.Process(ctx, tx, domain.NOTE, id)
	read.Run()
	if read.HasErr() {
		t.Fatalf("read: %v", read.GetErr())
	}
	return read.GetResult().Content
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
