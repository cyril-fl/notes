package millflagseditfolder

import (
	"context"
	"errors"
	"main.go/app/shared/services/testaccount"
	"testing"

	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createfolder"
	"main.go/app/shared/services/process/readitem"
	"main.go/config"
)

func TestResolve(t *testing.T) {
	ctx := context.Background()
	env := postgrey.NewEnvironment()
	env.ReadEnv()
	db := env.Connect()
	defer db.Close()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	ctx, tx = testaccount.New(t, ctx, tx)

	seed := createfolder.Process(ctx, tx, createfolder.Input{Title: "Old"})
	seed.Run()
	if seed.HasErr() {
		t.Fatalf("seed: %v", seed.GetErr())
	}

	cmd := &cobra.Command{}
	plumcliutilsargs.Makeflags(cmd, config.Current.Commands.EditFolder.Flags)
	if err := cmd.ParseFlags([]string{"--title", "New"}); err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	resolver := New(Template{Ctx: ctx, Db: tx, Flags: cmd.Flags(), ID: seed.GetResult().ID})
	if input := resolver.Resolve(); resolver.HasErr() || input.Title != "New" || input.ID != seed.GetResult().ID {
		t.Errorf("input = %+v, err = %v; want New", input, resolver.GetErr())
	}

	// Last: an unknown id fails before any form.
	missing := New(Template{Ctx: ctx, Db: tx, Flags: cmd.Flags(), ID: "00000000-0000-0000-0000-000000000000"})
	missing.Resolve()
	if !errors.Is(missing.GetErr(), readitem.ErrNotFound) {
		t.Errorf("err = %v, want %v", missing.GetErr(), readitem.ErrNotFound)
	}
}
