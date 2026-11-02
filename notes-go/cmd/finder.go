package cmd

import (
	"github.com/spf13/cobra"
	cmdfinder "main.go/app/cli/cmd-finder"
	"main.go/app/shared/domain"
)

var cmdFinder = &cobra.Command{
	Use:     cmd.ShowFinder.Use,
	Short:   cmd.ShowFinder.Short,
	Long:    cmd.ShowFinder.Long,
	Example: cmd.ShowFinder.Example,
	Args:    cobra.MaximumNArgs(1),
	RunE:    resolved(domain.FOLDER, cmdfinder.Finder),
}

func init() {
	cmdRoot.AddCommand(cmdFinder)
}
