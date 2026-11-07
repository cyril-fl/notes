package purgefiles

import (
	"context"
	"testing"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createfile"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/deleteitems"
	"main.go/app/shared/services/silo"
	"main.go/app/shared/services/testaccount"
)

func TestPurgeFiles(t *testing.T) {
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
	store := silo.NewMemory()

	// cited by a live note, cited by a trashed one, cited by none (old), cited by none (just pasted).
	files := map[string]domain.File{}
	for _, name := range []string{"live", "trashed", "orphan", "fresh"} {
		create := createfile.Process(ctx, tx, store, createfile.Input{Name: name, Data: []byte("\x89PNG\r\n\x1a\n")})
		create.Run()
		if create.HasErr() {
			t.Fatal(create.GetErr())
		}
		files[name] = create.GetResult()
	}
	var notes []string
	for _, name := range []string{"live", "trashed"} {
		note := createnote.Process(ctx, tx, createnote.Input{Content: "# " + name + "\n\n![](" + domain.FileURL(files[name].ID) + ")"})
		note.Run()
		if note.HasErr() {
			t.Fatal(note.GetErr())
		}
		notes = append(notes, note.GetResult().ID)
	}
	trash := deleteitems.Process(ctx, tx, domain.NOTE, notes[1:])
	trash.Run()
	if trash.HasErr() {
		t.Fatal(trash.GetErr())
	}
	_, err = tx.NewUpdate().Model((*domain.File)(nil)).Set("created_at = created_at - interval '2 hours'").
		Where("id != ? AND user_id = ?", files["fresh"].ID, domain.UserID(ctx)).Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	purge := Process(ctx, tx, store)
	purge.Run()
	if purge.HasErr() {
		t.Fatalf("purge: %v", purge.GetErr())
	}

	var left []string
	if err := tx.NewSelect().Model((*domain.File)(nil)).Column("name").Where("user_id = ?", domain.UserID(ctx)).Order("name").Scan(ctx, &left); err != nil {
		t.Fatal(err)
	}
	if got := len(left); got != 3 || left[0] != "fresh" || left[1] != "live" || left[2] != "trashed" {
		t.Errorf("left = %v, want [fresh live trashed]", left)
	}
	if _, ok := store.Objects[files["orphan"].Key()]; ok || len(store.Objects) != 3 {
		t.Errorf("objects = %v, want the orphan removed", store.Objects)
	}
}
