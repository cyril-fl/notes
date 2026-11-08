package cmd

import (
	cmdserve "github.com/cyril-fl/notes/notes-go/app/cli/cmd-serve"
	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
)

var cmdServe = &cobra.Command{
	Use:     cmd.Serve.Use,
	Short:   cmd.Serve.Short,
	Long:    cmd.Serve.Long,
	Example: cmd.Serve.Example,
	Args:    cobra.NoArgs,
	RunE:    cmdserve.Serve,
}

func init() {
	plumcliutilsargs.Makeflags(cmdServe, cmd.Serve.Flags)

	cmdRoot.AddCommand(cmdServe)
}
