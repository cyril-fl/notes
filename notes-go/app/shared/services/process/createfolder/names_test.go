// External package: the processes moving and restoring folders are tested here together, by name.
package createfolder_test

import (
	"context"
	"errors"
	"fmt"
	"main.go/app/shared/services/testaccount"
	"testing"
	"time"

	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createfolder"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/deleteitems"
	"main.go/app/shared/services/process/editfolder"
	"main.go/app/shared/services/process/moveitems"
	"main.go/app/shared/services/process/restoreitems"
)

func TestFolderNames(t *testing.T) {
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

	u := fmt.Sprintf("test-%d", time.Now().UnixNano())
	parent := folder(t, ctx, tx, u, "")
	a := folder(t, ctx, tx, "A", parent)
	b := folder(t, ctx, tx, "B", parent)
	elsewhere := folder(t, ctx, tx, "A", folder(t, ctx, tx, u+"-other", "")) // Same name, other folder: allowed.

	// Each refusal runs in a savepoint: a failed statement would abort the whole test transaction.
	taken := func(name string, run func(tx bun.Tx) error) {
		t.Helper()
		if err := tx.RunInTx(ctx, nil, func(_ context.Context, sub bun.Tx) error { return run(sub) }); !errors.Is(err, domain.ErrFolderExists) {
			t.Errorf("%s: err = %v, want %v", name, err, domain.ErrFolderExists)
		}
	}

	taken("create, other case", func(sub bun.Tx) error {
		p := createfolder.Process(ctx, sub, createfolder.Input{Title: "a", ParentID: parent})
		p.Run()
		return p.GetErr()
	})
	taken("rename", func(sub bun.Tx) error {
		p := editfolder.Process(ctx, sub, editfolder.Input{ID: b, Title: "A"})
		p.Run()
		return p.GetErr()
	})
	taken("move", func(sub bun.Tx) error {
		p := moveitems.Process(ctx, sub, moveitems.Input{IDs: []string{elsewhere}, ParentID: parent})
		p.Run()
		return p.GetErr()
	})

	// A trashed folder frees its name, and takes it back only if still free.
	trash := deleteitems.Process(ctx, tx, domain.FOLDER, []string{a})
	trash.Run()
	if trash.HasErr() {
		t.Fatalf("trash: %v", trash.GetErr())
	}
	folder(t, ctx, tx, "A", parent)
	taken("restore", func(sub bun.Tx) error {
		p := restoreitems.Process(ctx, sub, []string{a})
		p.Run()
		return p.GetErr()
	})
}

func folder(t *testing.T, ctx context.Context, tx bun.Tx, title, parentID string) string {
	t.Helper()
	process := createfolder.Process(ctx, tx, createfolder.Input{Title: title, ParentID: parentID})
	process.Run()
	if process.HasErr() {
		t.Fatalf("folder %s: %v", title, process.GetErr())
	}
	return process.GetResult().ID
}
