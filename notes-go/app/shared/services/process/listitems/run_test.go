package listitems

import (
	"context"
	"errors"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
	"slices"
	"testing"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createfolder"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/deleteitems"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/editfolder"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/moveitems"
)

func TestListItems(t *testing.T) {
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

	parent := createfolder.Process(ctx, tx, createfolder.Input{Title: "Parent"})
	parent.Run()
	if parent.HasErr() {
		t.Fatalf("parent: %v", parent.GetErr())
	}
	id := parent.GetResult().ID

	note := createnote.Process(ctx, tx, createnote.Input{Content: "# A note", ParentID: id})
	note.Run()
	if _, err := tx.NewRaw("UPDATE metadata_notes SET readonly = true WHERE item_id = ?", note.GetResult().ID).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	sub := createfolder.Process(ctx, tx, createfolder.Input{Title: "Z folder", ParentID: id})
	sub.Run()
	trashed := createfolder.Process(ctx, tx, createfolder.Input{Title: "Trashed", ParentID: id})
	trashed.Run()
	// Z folder's only subfolder is in the trash: it has none to show.
	gone := createfolder.Process(ctx, tx, createfolder.Input{Title: "Gone", ParentID: sub.GetResult().ID})
	gone.Run()
	trash := deleteitems.Process(ctx, tx, domain.FOLDER, []string{trashed.GetResult().ID, gone.GetResult().ID})
	trash.Run()
	if note.HasErr() || sub.HasErr() || trashed.HasErr() || gone.HasErr() || trash.HasErr() {
		t.Fatalf("seed: %v %v %v %v %v", note.GetErr(), sub.GetErr(), trashed.GetErr(), gone.GetErr(), trash.GetErr())
	}

	// Folders first even when a note sorts before them by title; the trashed folder is gone.
	children := Process(ctx, tx, id, "")
	children.Run()
	var got []string
	for _, item := range children.GetResult() {
		got = append(got, string(item.Title))
	}
	if children.HasErr() || !slices.Equal(got, []string{"Z folder", "A note"}) {
		t.Errorf("children = %v, %v; want [Z folder, A note]", got, children.GetErr())
	}
	if items := children.GetResult(); len(items) == 2 && (items[0].Readonly || !items[1].Readonly) {
		t.Errorf("readonly = %v, %v; want only the locked note", items[0].Readonly, items[1].Readonly)
	}
	if items := children.GetResult(); len(items) == 2 && (items[0].HasFolders || items[1].HasFolders) {
		t.Errorf("has_folders = %v, %v; want false (subfolder in the trash, a note)", items[0].HasFolders, items[1].HasFolders)
	}

	folders := Process(ctx, tx, id, domain.FOLDER)
	folders.Run()
	if items := folders.GetResult(); folders.HasErr() || len(items) != 1 || items[0].Title != "Z folder" {
		t.Errorf("folders = %v, %v; want only Z folder", items, folders.GetErr())
	}

	root := Process(ctx, tx, "", "")
	root.Run()
	found := slices.ContainsFunc(root.GetResult(), func(item domain.Item) bool { return item.ID == id && item.HasFolders })
	if root.HasErr() || !found {
		t.Errorf("root = %v; want Parent in it, with folders", root.GetErr())
	}

	kind := Process(ctx, tx, "", "tag")
	kind.Run()
	if !errors.Is(kind.GetErr(), ErrInvalidKind) {
		t.Errorf("err = %v, want %v", kind.GetErr(), ErrInvalidKind)
	}

	// Pages: the same order as the whole list, even under equal titles; the total counts every child.
	for range 3 {
		twin := createnote.Process(ctx, tx, createnote.Input{Content: "# Twin", ParentID: id})
		twin.Run()
		if twin.HasErr() {
			t.Fatalf("twin: %v", twin.GetErr())
		}
	}
	all := Process(ctx, tx, id, "")
	all.Run()
	var paged, want []string
	for _, item := range all.GetResult() {
		want = append(want, item.ID)
	}
	for offset := 0; offset < 6; offset += 2 {
		page := Process(ctx, tx, id, "").Page(2, offset)
		page.Run()
		if page.HasErr() || page.GetTotal() != 5 {
			t.Fatalf("page %d: total = %d, %v; want 5", offset, page.GetTotal(), page.GetErr())
		}
		for _, item := range page.GetResult() {
			paged = append(paged, item.ID)
		}
	}
	if all.HasErr() || len(want) != 5 || !slices.Equal(paged, want) {
		t.Errorf("pages = %v, want %v (%v)", paged, want, all.GetErr())
	}
	past := Process(ctx, tx, id, "").Page(2, 10)
	past.Run()
	if past.HasErr() || len(past.GetResult()) != 0 || past.GetTotal() != 5 {
		t.Errorf("past the end = %v, total %d, %v; want none of 5", past.GetResult(), past.GetTotal(), past.GetErr())
	}
	negative := Process(ctx, tx, id, "").Page(-1, 0)
	negative.Run()
	if !errors.Is(negative.GetErr(), ErrInvalidPage) {
		t.Errorf("err = %v, want %v", negative.GetErr(), ErrInvalidPage)
	}

	// The order is cached: reversed by hand, the next page follows it, until a write changes items_version.
	reverse := func() {
		cached, _ := orders.Load(domain.UserID(ctx))
		account := cached.(*cachedOrders)
		ids := slices.Clone(account.ids[orderKey{id, ""}])
		slices.Reverse(ids)
		account.ids[orderKey{id, ""}] = ids
	}
	first := func() (string, int) {
		page := Process(ctx, tx, id, "").Page(1, 0)
		page.Run()
		if page.HasErr() || len(page.GetResult()) != 1 {
			t.Fatalf("first: %v, %v", page.GetResult(), page.GetErr())
		}
		return string(page.GetResult()[0].Title), page.GetTotal()
	}
	reverse()
	if title, _ := first(); title != "Twin" {
		t.Errorf("first = %q, want Twin (the reversed order in cache)", title)
	}
	added := createnote.Process(ctx, tx, createnote.Input{Content: "# Added", ParentID: id})
	writes := []struct {
		name  string
		write interface {
			Run()
			HasErr() bool
			GetErr() error
		}
		title string
		total int
	}{
		{"add", added, "Z folder", 6},
		{"rename", editfolder.Process(ctx, tx, editfolder.Input{ID: sub.GetResult().ID, Title: "Y folder"}), "Y folder", 6},
		{"move", nil, "Y folder", 5},
		{"trash", deleteitems.Process(ctx, tx, domain.FOLDER, []string{sub.GetResult().ID}), "A note", 4},
	}
	for _, w := range writes {
		first()
		reverse()
		if w.write == nil { // after add ran: its note is known
			w.write = moveitems.Process(ctx, tx, moveitems.Input{IDs: []string{added.GetResult().ID}})
		}
		w.write.Run()
		if w.write.HasErr() {
			t.Fatalf("%s: %v", w.name, w.write.GetErr())
		}
		if title, total := first(); title != w.title || total != w.total {
			t.Errorf("after %s: first = %q of %d, want %q of %d", w.name, title, total, w.title, w.total)
		}
	}

	// Last: a DB error.
	invalid := Process(ctx, tx, "nope", "")
	invalid.Run()
	if !errors.Is(invalid.GetErr(), ErrInvalidID) {
		t.Errorf("err = %v, want %v", invalid.GetErr(), ErrInvalidID)
	}
}
