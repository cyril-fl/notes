package millflagscreatefolder

import (
	"context"
	"testing"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createfolder"
	"github.com/cyril-fl/notes/notes-go/config"
	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
)

// resolve parses args with the flags declared in config.yaml: a renamed flag breaks this test.
// Every case gives a title, so no form opens and no terminal is needed.
func resolve(t *testing.T, args ...string) (*mill, createfolder.Input) {
	t.Helper()
	cmd := &cobra.Command{}
	plumcliutilsargs.Makeflags(cmd, config.Current.Commands.CreateFolder.Flags)
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	resolver := New(Template{Ctx: context.Background(), Flags: cmd.Flags()})
	return resolver, resolver.Resolve()
}

func TestResolveFromFlags(t *testing.T) {
	resolver, input := resolve(t, "--title", "T", "--parent", "uuid")
	if resolver.HasErr() || input.Title != "T" || input.ParentID != "uuid" {
		t.Errorf("input = %+v, err = %v; want T in uuid", input, resolver.GetErr())
	}
}
