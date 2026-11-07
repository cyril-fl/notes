package deleteitems

import (
	"context"
	"errors"
	"main.go/app/shared/services/testaccount"
	"testing"
	"time"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createfolder"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
)

// No DB needed: makeIDs fails first and markDeleted is guarded, so a nil db is never touched.
func TestBlankValues(t *testing.T) {
	process := Process(context.Background(), nil, domain.NOTE, []string{"", "  "})
	process.Run()

	if !errors.Is(process.GetErr(), ErrNoValue) {
		t.Fatalf("err = %v, want %v", process.GetErr(), ErrNoValue)
	}
}

func TestDeleteFolderTree(t *testing.T) {
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

	folder := func(parentID string) string {
		p := createfolder.Process(ctx, tx, createfolder.Input{Title: "folder", ParentID: parentID})
		p.Run()
		if p.HasErr() {
			t.Fatalf("folder: %v", p.GetErr())
		}
		return p.GetResult().ID
	}
	note := func(parentID string) string {
		p := createnote.Process(ctx, tx, createnote.Input{Content: "content", ParentID: parentID})
		p.Run()
		if p.HasErr() {
			t.Fatalf("note: %v", p.GetErr())
		}
		return p.GetResult().ID
	}

	root := folder("")
	child := folder(root)
	nested := note(child)
	alone := note("")

	run := func(kind domain.Kind, ids ...string) []domain.Item {
		p := Process(ctx, tx, kind, ids)
		p.Run()
		if p.HasErr() {
			t.Fatalf("delete %s %v: %v", kind, ids, p.GetErr())
		}
		return p.GetResult()
	}

	if got := run(domain.NOTE, root); len(got) != 0 {
		t.Errorf("remove note <folder> = %+v, want nothing", got)
	}
	if got := run(domain.FOLDER, root, "00000000-0000-0000-0000-000000000000"); len(got) != 3 {
		t.Errorf("remove folder <root> = %+v, want root, child and nested note", got)
	}
	if got := run(domain.FOLDER, root); len(got) != 0 {
		t.Errorf("second remove = %+v, want nothing", got)
	}

	// bun hides soft-deleted rows: only the lone note is still visible.
	visible, err := tx.NewSelect().Model((*domain.Item)(nil)).Where("id IN (?, ?, ?, ?)", root, child, nested, alone).Count(ctx)
	if err != nil || visible != 1 {
		t.Errorf("visible = %d, %v; want 1", visible, err)
	}

	var dates []time.Time
	err = tx.NewSelect().Model((*domain.Item)(nil)).WhereDeleted().Column("deleted_at").Where("id IN (?, ?, ?)", root, child, nested).Scan(ctx, &dates)
	if err != nil || len(dates) != 3 || !dates[0].Equal(dates[1]) || !dates[0].Equal(dates[2]) {
		t.Errorf("deleted_at = %v, %v; want 3 equal dates", dates, err)
	}

	if got := run(domain.NOTE, alone); len(got) != 1 {
		t.Errorf("remove note <alone> = %+v, want it", got)
	}

	// Last: Postgres rejects a malformed uuid and aborts the transaction.
	invalid := Process(ctx, tx, domain.NOTE, []string{"not-a-uuid"})
	invalid.Run()
	if !invalid.HasErr() {
		t.Errorf("invalid uuid: err = nil, want a Postgres error")
	}
}

// A [[link]] waiting because two notes shared its title resolves when one goes to the trash.
func TestTrashResolvesLinks(t *testing.T) {
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

	note := func(content string) string {
		process := createnote.Process(ctx, tx, createnote.Input{Content: content})
		process.Run()
		if process.HasErr() {
			t.Fatalf("note %q: %v", content, process.GetErr())
		}
		return process.GetResult().ID
	}

	trashed := note("# Bee")
	kept := note("# Bee")
	a := note("see [[Bee]]")

	remove := Process(ctx, tx, domain.NOTE, []string{trashed})
	remove.Run()
	if remove.HasErr() {
		t.Fatal(remove.GetErr())
	}

	var content string
	if err := tx.NewRaw("SELECT content FROM metadata_notes WHERE item_id = ?", a).Scan(ctx, &content); err != nil {
		t.Fatal(err)
	}
	if want := "see [Bee](note:" + kept + ")"; content != want {
		t.Errorf("content = %q, want %q", content, want)
	}
}
