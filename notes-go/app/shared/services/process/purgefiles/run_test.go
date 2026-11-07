package purgefiles

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createfile"
	"main.go/app/shared/services/process/createfolder"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/deleteitems"
	"main.go/app/shared/services/process/emptytrash"
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

	// cited by a live note, by a trashed one, by none for long, by none since just now, cited again, the avatar.
	names := []string{"live", "trashed", "orphan", "removed", "back", "avatar"}
	files := map[string]domain.File{}
	for _, name := range names {
		create := createfile.Process(ctx, tx, store, createfile.Input{Name: name, Data: []byte("\x89PNG\r\n\x1a\n")})
		create.Run()
		if create.HasErr() {
			t.Fatal(create.GetErr())
		}
		files[name] = create.GetResult()
	}
	var notes []string
	for _, name := range []string{"live", "trashed", "back"} {
		note := createnote.Process(ctx, tx, createnote.Input{Content: "# " + name + "\n\n![](" + domain.FileURL(files[name].ID) + ")"})
		note.Run()
		if note.HasErr() {
			t.Fatal(note.GetErr())
		}
		notes = append(notes, note.GetResult().ID)
	}
	trash := deleteitems.Process(ctx, tx, domain.NOTE, notes[1:2])
	trash.Run()
	if trash.HasErr() {
		t.Fatal(trash.GetErr())
	}
	// The account keeps its trash 30 days: orphan, back and avatar were left uncited 40 days ago.
	_, err = tx.NewUpdate().Model((*domain.File)(nil)).Set("orphaned_at = current_timestamp - interval '40 days'").
		Where("id IN (?, ?, ?)", files["orphan"].ID, files["back"].ID, files["avatar"].ID).Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.NewUpdate().Model((*domain.User)(nil)).Set("avatar = ?", files["avatar"].ID).Where("id = ?", domain.UserID(ctx)).Exec(ctx); err != nil {
		t.Fatal(err)
	}

	purge := Process(ctx, tx, store)
	purge.Run()
	if purge.HasErr() {
		t.Fatalf("purge: %v", purge.GetErr())
	}

	var left []domain.File
	if err := tx.NewSelect().Model(&left).Where("user_id = ?", domain.UserID(ctx)).Order("name").Scan(ctx); err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, file := range left {
		got = append(got, fmt.Sprintf("%s:%t", file.Name, file.OrphanedAt.IsZero()))
	}
	// removed waits, its clock started; back is cited again and avatar is the account's, their clocks stopped.
	if want := "avatar:true back:true live:true removed:false trashed:true"; strings.Join(got, " ") != want {
		t.Errorf("left = %v, want %s", got, want)
	}
	if _, ok := store.Objects[files["orphan"].Key()]; ok || len(store.Objects) != 5 {
		t.Errorf("objects = %v, want the orphan removed", store.Objects)
	}

	// With a retention of 0, nothing goes.
	if _, err := tx.NewUpdate().Model((*domain.User)(nil)).Set("trash_retention_days = 0").Where("id = ?", domain.UserID(ctx)).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	_, err = tx.NewUpdate().Model((*domain.File)(nil)).Set("orphaned_at = current_timestamp - interval '400 days'").
		Where("id = ?", files["removed"].ID).Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	keep := Process(ctx, tx, store)
	keep.Run()
	if keep.HasErr() || keep.GetResult() != 0 {
		t.Errorf("purge with no retention = %d, %v, want 0", keep.GetResult(), keep.GetErr())
	}

	// Notes purged for good, here one by cascade with its folder, frees at once what it alone cited, retention or not.
	folder := createfolder.Process(ctx, tx, createfolder.Input{Title: fmt.Sprintf("purge-%d", time.Now().UnixNano())})
	folder.Run()
	if folder.HasErr() {
		t.Fatal(folder.GetErr())
	}
	content := "# purged\n\n![](" + domain.FileURL(files["removed"].ID) + ") ![](" + domain.FileURL(files["live"].ID) + ")"
	purged := createnote.Process(ctx, tx, createnote.Input{Content: content, ParentID: folder.GetResult().ID})
	purged.Run()
	if purged.HasErr() {
		t.Fatal(purged.GetErr())
	}
	for _, step := range []interface {
		Run()
		HasErr() bool
		GetErr() error
	}{
		deleteitems.Process(ctx, tx, domain.FOLDER, []string{folder.GetResult().ID}),
		emptytrash.Process(ctx, tx),
	} {
		step.Run()
		if step.HasErr() {
			t.Fatal(step.GetErr())
		}
	}
	freed := Process(ctx, tx, store)
	freed.Run()
	// The whole trash went: the note of trashed too.
	if freed.HasErr() || freed.GetResult() != 2 {
		t.Errorf("purge after the trash = %d, %v, want removed and trashed (live is cited elsewhere)", freed.GetResult(), freed.GetErr())
	}
	if _, ok := store.Objects[files["live"].Key()]; !ok {
		t.Error("live was removed, a live note cites it")
	}
	var orphaned int
	if err := tx.NewRaw("SELECT count(*) FROM files WHERE id = ? AND orphaned_at IS NULL", files["live"].ID).Scan(ctx, &orphaned); err != nil || orphaned != 1 {
		t.Errorf("live: orphaned_at not cleared (%d, %v)", orphaned, err)
	}
}
