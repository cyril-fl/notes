package millflagscreatenote

import (
	"context"
	"slices"
	"testing"

	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
	"main.go/app/shared/services/process/createnote"
	"main.go/config"
)

// resolve parses args with the flags declared in config.yaml: a renamed flag breaks this test.
// Every case gives title and content, so no form opens and no terminal is needed.
func resolve(t *testing.T, args ...string) (*mill, createnote.Input) {
	t.Helper()
	cmd := &cobra.Command{}
	plumcliutilsargs.Makeflags(cmd, config.Current.Commands.CreateNote.Flags)
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	resolver := New(Template{Ctx: context.Background(), Flags: cmd.Flags()})
	return resolver, resolver.Resolve()
}

func TestResolveFromFlags(t *testing.T) {
	resolver, input := resolve(t, "--title", "T", "--content", "C", "--tag", "go,sql", "--tag", "db", "--parent", "uuid")
	if resolver.HasErr() {
		t.Fatalf("Resolve() error = %v", resolver.GetErr())
	}

	if input.Title != "T" || input.Content != "C" || input.ParentID != "uuid" || !slices.Equal(input.Tags, []string{"go", "sql", "db"}) {
		t.Errorf("input = %+v", input)
	}
}

func TestResolveNoTitle(t *testing.T) {
	resolver, input := resolve(t, "--no-title", "--content", "C")
	if resolver.HasErr() || input.Title != "" || input.Content != "C" {
		t.Errorf("input = %+v, err = %v; want untitled C", input, resolver.GetErr())
	}
}
