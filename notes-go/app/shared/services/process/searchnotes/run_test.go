package searchnotes

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
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/deleteitems"
)

// No DB needed: makeText fails first and the search is guarded, so a nil db is never touched.
func TestNoValue(t *testing.T) {
	process := Process(context.Background(), nil, " ")
	process.Run()
	if !errors.Is(process.GetErr(), ErrNoValue) {
		t.Fatalf("err = %v, want %v", process.GetErr(), ErrNoValue)
	}
}

func TestSnippet(t *testing.T) {
	long := strings.Repeat("a", 80) + " needle " + strings.Repeat("b", 80)
	tests := map[string]struct{ content, text, want string }{
		"line found":   {"first\nthe NEEDLE here\nlast", "needle", "the NEEDLE here"},
		"title only":   {"  first line  \nsecond", "absent", "first line"},
		"cut around":   {long, "needle", "…" + strings.Repeat("a", 19) + " needle " + strings.Repeat("b", 33) + "…"},
		"short enough": {"x needle y", "needle", "x needle y"},
	}

	for name, test := range tests {
		if got := snippet(test.content, test.text); got != test.want {
			t.Errorf("%s: snippet() = %q, want %q", name, got, test.want)
		}
	}
}

func TestSearchNotes(t *testing.T) {
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

	// Unique per run so real notes cannot match.
	word := fmt.Sprintf("zq%d", time.Now().UnixNano())
	ids := map[string]string{}
	for title, content := range map[string]string{
		"A in content":    "line\nhas " + strings.ToUpper(word) + " inside",
		"Z " + word:       "title match",
		"Percent":         word + "100% sure",
		"Trashed " + word: "C",
	} {
		process := createnote.Process(ctx, tx, createnote.Input{Content: "# " + title + "\n\n" + content})
		process.Run()
		if process.HasErr() {
			t.Fatalf("note %s: %v", title, process.GetErr())
		}
		ids[title] = process.GetResult().ID
	}
	trash := deleteitems.Process(ctx, tx, domain.NOTE, []string{ids["Trashed "+word]})
	trash.Run()
	if trash.HasErr() {
		t.Fatalf("trash: %v", trash.GetErr())
	}

	// Title match first, then by path; case-insensitive; the trashed note is gone.
	search := Process(ctx, tx, word)
	search.Run()
	var titles []string
	for _, match := range search.GetResult() {
		titles = append(titles, match.Title)
	}
	if search.HasErr() || strings.Join(titles, ",") != "Z "+word+",A in content,Percent" {
		t.Fatalf("titles = %v, %v", titles, search.GetErr())
	}
	if got := search.GetResult()[1].Snippet; got != "has "+strings.ToUpper(word)+" inside" {
		t.Errorf("snippet = %q, want the matching line", got)
	}

	// % is literal: only the note containing "100%" matches.
	percent := Process(ctx, tx, word+"100%")
	percent.Run()
	if percent.HasErr() || len(percent.GetResult()) != 1 || percent.GetResult()[0].Title != "Percent" {
		t.Errorf("percent = %+v, %v; want only Percent", percent.GetResult(), percent.GetErr())
	}
}
