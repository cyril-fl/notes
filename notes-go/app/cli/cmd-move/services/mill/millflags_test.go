package millflagsmove

import (
	"slices"
	"testing"

	"github.com/cyril-fl/notes/notes-go/config"
	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
)

// resolve parses args with the flags declared in config.yaml: a renamed flag breaks this test.
func resolve(t *testing.T, args ...string) *mill {
	t.Helper()

	cmd := &cobra.Command{}
	plumcliutilsargs.Makeflags(cmd, config.Current.Commands.Move.Flags)
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	return New(Template{Flags: cmd.Flags(), IDs: cmd.Flags().Args()})
}

func TestResolve(t *testing.T) {
	resolver := resolve(t, "a", "b", "--parent", "p")
	input := resolver.Resolve()
	if resolver.HasErr() || input.ParentID != "p" || !slices.Equal(input.IDs, []string{"a", "b"}) {
		t.Errorf("input = %+v, err = %v", input, resolver.GetErr())
	}
}

func TestResolveRoot(t *testing.T) {
	resolver := resolve(t, "a")
	if input := resolver.Resolve(); resolver.HasErr() || input.ParentID != "" {
		t.Errorf("input = %+v, err = %v; want the root", input, resolver.GetErr())
	}
}
