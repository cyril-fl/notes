package cmd

import (
	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
	cmdsearch "main.go/app/cli/cmd-search"
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
