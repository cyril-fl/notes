package cmd

import (
	cmdrecent "github.com/cyril-fl/notes/notes-go/app/cli/cmd-recent"
	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
)

var cmdRecent = &cobra.Command{
	Use:     cmd.Recent.Use,
	Short:   cmd.Recent.Short,
	Long:    cmd.Recent.Long,
	Example: cmd.Recent.Example,
	Args:    cobra.NoArgs,
	RunE:    cmdrecent.Recent,
}

func init() {
	plumcliutilsargs.Makeflags(cmdRecent, cmd.Recent.Flags)

	cmdRoot.AddCommand(cmdRecent)
}
