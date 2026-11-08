package createnote

import (
	"context"
	"errors"
	"fmt"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
	"strings"
	"testing"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/hashtag"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createtag"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
)

// No DB needed: makeNote fails first and saveNote is guarded, so a nil db is never touched.
func TestNoContent(t *testing.T) {
	process := Process(context.Background(), nil, Input{Content: "   "})
	process.Run()

	if !errors.Is(process.GetErr(), ErrNoContent) {
		t.Fatalf("err = %v, want %v", process.GetErr(), ErrNoContent)
	}
}

func TestCreateNote(t *testing.T) {
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
	existing := fmt.Sprintf("test-%d", time.Now().UnixNano())
	seed := createtag.Process(ctx, tx, []string{existing})
	seed.Run()
	if seed.HasErr() {
		t.Fatalf("seed: %v", seed.GetErr())
	}

	// Existing tag, new tag, and a case duplicate of the existing one: 2 links.
	note := Process(ctx, tx, Input{Content: "# Title \n\ncontent", Tags: []string{existing, existing + "-new", strings.ToUpper(existing), " "}})
	note.Run()
	if note.HasErr() {
		t.Fatalf("create: %v", note.GetErr())
	}
	item := note.GetResult()
	if item.ID == "" || item.Title != "Title" || item.Kind != domain.NOTE {
		t.Fatalf("result = %+v, want an id, Title, note", item)
	}

	links, err := tx.NewSelect().Model((*domain.NoteTag)(nil)).Where("note_id = ?", item.ID).Count(ctx)
	if err != nil || links != 2 {
		t.Errorf("links = %d, %v; want 2", links, err)
	}

	var content secret.Text
	err = tx.NewSelect().Model((*domain.MetadataNote)(nil)).Column("content").Where("item_id = ?", item.ID).Scan(ctx, &content)
	if err != nil || !strings.HasPrefix(string(content), "# Title \n\ncontent\n\n#") {
		t.Errorf("content = %q, %v; want content then its tags", content, err)
	}

	// The base keeps the title and the tags sealed: nothing of them in clear.
	var clear int
	err = tx.NewRaw("SELECT count(*) FROM items AS i, tags AS t WHERE i.id = ? AND t.user_id = i.user_id AND (i.title NOT LIKE 'enc1:%' OR t.value NOT LIKE 'enc1:%' OR t.value ILIKE ?)",
		item.ID, "%"+existing+"%").Scan(ctx, &clear)
	if err != nil || clear != 0 {
		t.Errorf("rows in clear = %d, %v; want 0", clear, err)
	}

	parsed := Process(ctx, tx, Input{Content: "avec #" + "testparsed"})
	parsed.Run()
	if parsed.HasErr() {
		t.Fatalf("parsed: %v", parsed.GetErr())
	}
	links, err = tx.NewSelect().Model((*domain.NoteTag)(nil)).Where("note_id = ?", parsed.GetResult().ID).Count(ctx)
	if err != nil || links != 1 {
		t.Errorf("links of a #tag = %d, %v; want 1", links, err)
	}

	invalid := Process(ctx, tx, Input{Content: "c", Tags: []string{"a b"}})
	invalid.Run()
	if !errors.Is(invalid.GetErr(), hashtag.ErrInvalid) {
		t.Errorf("invalid tag: err = %v, want %v", invalid.GetErr(), hashtag.ErrInvalid)
	}

	words := Process(ctx, tx, Input{Content: "- [ ] du **pain**\net du lait"})
	words.Run()
	if words.HasErr() || words.GetResult().Title != "du pain" {
		t.Errorf("title = %q, %v; want the first words", words.GetResult().Title, words.GetErr())
	}

	// Nothing to read a title from: only code.
	untitled := Process(ctx, tx, Input{Content: "```\ncode\n```"})
	untitled.Run()
	if untitled.HasErr() {
		t.Fatalf("untitled: %v", untitled.GetErr())
	}
	var isNull bool
	err = tx.NewSelect().Model((*domain.Item)(nil)).ColumnExpr("title IS NULL").Where("id = ?", untitled.GetResult().ID).Scan(ctx, &isNull)
	if err != nil || !isNull {
		t.Errorf("title IS NULL = %v, %v; want true", isNull, err)
	}

	// Last: the parent FK fails (RunInTx uses a savepoint here, but keep DB errors at the end).
	orphan := Process(ctx, tx, Input{Content: "content", ParentID: "00000000-0000-0000-0000-000000000000"})
	orphan.Run()
	if !errors.Is(orphan.GetErr(), ErrParentNotFound) {
		t.Errorf("err = %v, want %v", orphan.GetErr(), ErrParentNotFound)
	}
}
