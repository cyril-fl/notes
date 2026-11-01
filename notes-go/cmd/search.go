package cmd

import (
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
	cmdRoot.AddCommand(cmdSearch)
}
