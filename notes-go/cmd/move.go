package cmd

import (
	cmdmove "github.com/cyril-fl/notes/notes-go/app/cli/cmd-move"
	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
)

var cmdMove = &cobra.Command{
	Use:     cmd.Move.Use,
	Short:   cmd.Move.Short,
	Long:    cmd.Move.Long,
	Example: cmd.Move.Example,
	Args:    cobra.MinimumNArgs(1),
	RunE:    resolved("", cmdmove.Move),
}

func init() {
	plumcliutilsargs.Makeflags(cmdMove, cmd.Move.Flags)
	cmdMove.MarkFlagsMutuallyExclusive("parent", "parent-from-list")

	cmdRoot.AddCommand(cmdMove)
}
