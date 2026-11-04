package millflagseditnote

import (
	"context"
	"slices"
	"testing"

	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
	"main.go/config"
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

	seed := createnote.Process(ctx, tx, createnote.Input{Title: "Old", Content: "old", Tags: []string{"test-edit-a", "test-edit-b"}})
	seed.Run()
	if seed.HasErr() {
		t.Fatalf("seed: %v", seed.GetErr())
	}

	cmd := &cobra.Command{}
	plumcliutilsargs.Makeflags(cmd, config.Current.Commands.EditNote.Flags)
	if err := cmd.ParseFlags([]string{"--title", "", "--tag", "test-edit-c", "--untag", "TEST-EDIT-A"}); err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	resolver := New(Template{Ctx: ctx, Db: tx, Flags: cmd.Flags(), ID: seed.GetResult().ID})
	input := resolver.Resolve()
	if resolver.HasErr() {
		t.Fatalf("Resolve() error = %v", resolver.GetErr())
	}
	// --title "" untitles, content is kept (editnote writes the tags into it), tags are merged.
	if input.Title != "" || input.Content != "old\n\n#test-edit-a #test-edit-b" || !slices.Equal(input.Tags, []string{"test-edit-b", "test-edit-c"}) {
		t.Errorf("input = %+v", input)
	}
}
