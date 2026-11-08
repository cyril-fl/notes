package cmd

import (
	cmdsearch "github.com/cyril-fl/notes/notes-go/app/cli/cmd-search"
	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
)

var cmdSearch = &cobra.Command{
	Use:     cmd.Search.Use,
	Short:   cmd.Search.Short,
	Long:    cmd.Search.Long,
	Example: cmd.Search.Example,
	Args:    cobra.MinimumNArgs(1),
	RunE:    cmdsearch.Search,
}

func init() {
	plumcliutilsargs.Makeflags(cmdSearch, cmd.Search.Flags)

	cmdRoot.AddCommand(cmdSearch)
}
