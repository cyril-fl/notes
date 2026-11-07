package purgeitems

import (
	"context"
	"main.go/app/shared/services/testaccount"
	"testing"

	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/deleteitems"
)

func TestPurgeOldOnly(t *testing.T) {
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

	ids := make([]string, 3) // live, deleted now, deleted 40 days ago
	for i := range ids {
		note := createnote.Process(ctx, tx, createnote.Input{Content: "content"})
		note.Run()
		if note.HasErr() {
			t.Fatalf("note: %v", note.GetErr())
		}
		ids[i] = note.GetResult().ID
	}

	trash := deleteitems.Process(ctx, tx, domain.NOTE, ids[1:])
	trash.Run()
	if trash.HasErr() {
		t.Fatalf("delete: %v", trash.GetErr())
	}
	_, err = tx.NewUpdate().Model((*domain.Item)(nil)).WhereAllWithDeleted().
		Set("deleted_at = deleted_at - interval '40 days'").Where("id = ?", ids[2]).Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	count := func() int64 {
		left, err := tx.NewSelect().Model((*domain.Item)(nil)).WhereAllWithDeleted().Where("id IN (?)", bun.In(ids)).Count(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return left
	}
	setRetention := func(days int) {
		_, err := tx.NewUpdate().Model((*domain.User)(nil)).Set("trash_retention_days = ?", days).Where("id = ?", domain.UserID(ctx)).Exec(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}

	setRetention(0)
	purge := Process(ctx, tx)
	purge.Run()
	if purge.HasErr() {
		t.Fatalf("purge: %v", purge.GetErr())
	}
	if left := count(); left != 3 {
		t.Errorf("left with retention 0 = %d, want 3: nothing purged", left)
	}

	setRetention(30)
	purge = Process(ctx, tx)
	purge.Run()
	if purge.HasErr() {
		t.Fatalf("purge: %v", purge.GetErr())
	}

	left, err := tx.NewSelect().Model((*domain.Item)(nil)).WhereAllWithDeleted().Where("id IN (?)", bun.In(ids)).Count(ctx)
	if err != nil || left != 2 {
		t.Errorf("left = %d, %v; want the live and the recently deleted notes", left, err)
	}
	metadata, err := tx.NewSelect().Model((*domain.MetadataNote)(nil)).Where("item_id = ?", ids[2]).Count(ctx)
	if err != nil || metadata != 0 {
		t.Errorf("metadata of purged note = %d, %v; want 0 (cascade)", metadata, err)
	}
}
