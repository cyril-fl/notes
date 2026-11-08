package cmd

import (
	cmdfinder "github.com/cyril-fl/notes/notes-go/app/cli/cmd-finder"
	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/spf13/cobra"
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
