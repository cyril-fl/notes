package cmd

import (
	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
	cmdrecent "main.go/app/cli/cmd-recent"
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
