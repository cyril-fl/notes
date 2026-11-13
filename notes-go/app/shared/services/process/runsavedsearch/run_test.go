package runsavedsearch

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createfolder"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createtag"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
)

func titles(matches []Match) []string {
	var got []string
	for _, m := range matches {
		got = append(got, m.Title)
	}
	return got
}

func TestRunSavedSearch(t *testing.T) {
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

	word := fmt.Sprintf("zq%d", time.Now().UnixNano())

	tags := createtag.Process(ctx, tx, []string{word + "-a", word + "-b"})
	tags.Run()
	if tags.HasErr() {
		t.Fatalf("tags: %v", tags.GetErr())
	}
	var tagA, tagB string
	for _, tag := range tags.GetResult() {
		switch string(tag.Value) {
		case word + "-a":
			tagA = tag.ID
		case word + "-b":
			tagB = tag.ID
		}
	}

	folder := createfolder.Process(ctx, tx, createfolder.Input{Title: "Folder " + word})
	folder.Run()
	if folder.HasErr() {
		t.Fatalf("folder: %v", folder.GetErr())
	}
	folderID := folder.GetResult().ID

	noteIn := createnote.Process(ctx, tx, createnote.Input{
		Content:  "# In folder, tag a\n\nneedle " + word,
		ParentID: folderID, Tags: []string{word + "-a"},
	})
	noteIn.Run()
	if noteIn.HasErr() {
		t.Fatalf("note in: %v", noteIn.GetErr())
	}

	noteBoth := createnote.Process(ctx, tx, createnote.Input{
		Content: "# Outside, both tags",
		Tags:    []string{word + "-a", word + "-b"},
	})
	noteBoth.Run()
	if noteBoth.HasErr() {
		t.Fatalf("note both: %v", noteBoth.GetErr())
	}

	noteOther := createnote.Process(ctx, tx, createnote.Input{Content: "# Outside, no tag\n\nunrelated"})
	noteOther.Run()
	if noteOther.HasErr() {
		t.Fatalf("note other: %v", noteOther.GetErr())
	}

	// Tags "any": the note in the folder and the note with both tags, not the untagged one.
	any := Process(ctx, tx, domain.SavedSearch{TagIDs: []string{tagA, tagB}, TagMode: "any"})
	any.Run()
	if any.HasErr() || len(any.GetResult()) != 2 {
		t.Fatalf("any = %+v, %v; want 2 matches", any.GetResult(), any.GetErr())
	}

	// Tags "all": only the note carrying both.
	all := Process(ctx, tx, domain.SavedSearch{TagIDs: []string{tagA, tagB}, TagMode: "all"})
	all.Run()
	if all.HasErr() || strings.Join(titles(all.GetResult()), ",") != "Outside, both tags" {
		t.Fatalf("all = %+v, %v; want only the note with both tags", all.GetResult(), all.GetErr())
	}

	// Folder included: only the note inside it.
	included := Process(ctx, tx, domain.SavedSearch{FolderIDs: []string{folderID}, FolderMode: "include"})
	included.Run()
	if included.HasErr() || strings.Join(titles(included.GetResult()), ",") != "In folder, tag a" {
		t.Fatalf("included = %+v, %v; want only the note in the folder", included.GetResult(), included.GetErr())
	}

	// Folder excluded: everything but the note inside it.
	excluded := Process(ctx, tx, domain.SavedSearch{FolderIDs: []string{folderID}, FolderMode: "exclude"})
	excluded.Run()
	if excluded.HasErr() || len(excluded.GetResult()) != 2 {
		t.Fatalf("excluded = %+v, %v; want 2 matches", excluded.GetResult(), excluded.GetErr())
	}

	// Text narrows the SQL-filtered candidates, not the whole account: "unrelated" only matches outside the tag filter.
	text := Process(ctx, tx, domain.SavedSearch{TagIDs: []string{tagA}, TagMode: "any", Text: "needle"})
	text.Run()
	if text.HasErr() || strings.Join(titles(text.GetResult()), ",") != "In folder, tag a" {
		t.Fatalf("text = %+v, %v; want only the note with the needle", text.GetResult(), text.GetErr())
	}
	if text.GetResult()[0].Snippet == "" {
		t.Errorf("snippet is empty, want the matching line")
	}

	// No criterion at all: every live note of the account.
	all3 := Process(ctx, tx, domain.SavedSearch{})
	all3.Run()
	if all3.HasErr() || len(all3.GetResult()) != 3 {
		t.Fatalf("no criteria = %+v, %v; want every note", all3.GetResult(), all3.GetErr())
	}
}
