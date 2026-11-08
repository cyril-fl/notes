package millflagseditnote

import (
	"context"
	"errors"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
	"slices"
	"testing"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/patchnote"
	"github.com/cyril-fl/notes/notes-go/config"
	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
)

func TestMergeTags(t *testing.T) {
	got := mergeTags([]string{"Go", "sql"}, []string{"go", "db", " ", "DB"}, []string{"SQL"})
	if want := []string{"Go", "db"}; !slices.Equal(got, want) {
		t.Errorf("mergeTags() = %v, want %v", got, want)
	}
}

// Flags come from config.yaml (a renamed flag breaks this test) and always give a change, so no form opens.
func TestResolveFromFlags(t *testing.T) {
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

	seed := createnote.Process(ctx, tx, createnote.Input{Content: "old", Tags: []string{"test-edit-a", "test-edit-b"}})
	seed.Run()
	if seed.HasErr() {
		t.Fatalf("seed: %v", seed.GetErr())
	}

	cmd := &cobra.Command{}
	plumcliutilsargs.Makeflags(cmd, config.Current.Commands.EditNote.Flags)
	if err := cmd.ParseFlags([]string{"--tag", "test-edit-c", "--untag", "TEST-EDIT-A"}); err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	resolver := New(Template{Ctx: ctx, Db: tx, Flags: cmd.Flags(), ID: seed.GetResult().ID})
	input := resolver.Resolve()
	if resolver.HasErr() {
		t.Fatalf("Resolve() error = %v", resolver.GetErr())
	}
	// Content is kept (editnote writes the tags into it), tags are merged.
	if input.Content != "old\n\n#test-edit-a #test-edit-b" || !slices.Equal(input.Tags, []string{"test-edit-b", "test-edit-c"}) {
		t.Errorf("input = %+v", input)
	}
	// The version read is the one patchnote checks: a save over it passes, over an older one is stale.
	// current_timestamp is frozen in the transaction, so a later edit cannot make the version old here.
	version := resolver.Version()
	older := version.Add(-time.Second)
	for _, tc := range []struct {
		version *time.Time
		want    error
	}{{&version, nil}, {&older, patchnote.ErrStale}} {
		save := patchnote.Process(ctx, tx, patchnote.Input{ID: input.ID, Content: &input.Content, Tags: &input.Tags, UpdatedAt: tc.version})
		save.Run()
		if !errors.Is(save.GetErr(), tc.want) {
			t.Errorf("save over %v: error = %v, want %v", tc.version, save.GetErr(), tc.want)
		}
	}
}
