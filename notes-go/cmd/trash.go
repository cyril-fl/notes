package cmd

import (
	cmdrestore "github.com/cyril-fl/notes/notes-go/app/cli/cmd-restore"
	cmdtrash "github.com/cyril-fl/notes/notes-go/app/cli/cmd-trash"
	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
)

var cmdTrash = &cobra.Command{
	Use:     cmd.Trash.Use,
	Short:   cmd.Trash.Short,
	Long:    cmd.Trash.Long,
	Example: cmd.Trash.Example,
	Args:    cobra.NoArgs,
	RunE:    cmdtrash.Trash,
}

var cmdRestore = &cobra.Command{
	Use:     cmd.Restore.Use,
	Short:   cmd.Restore.Short,
	Long:    cmd.Restore.Long,
	Example: cmd.Restore.Example,
	Args:    cobra.MinimumNArgs(1),
	RunE:    cmdrestore.Restore,
}

func init() {
	plumcliutilsargs.Makeflags(cmdTrash, cmd.Trash.Flags)

	cmdRoot.AddCommand(cmdTrash, cmdRestore)
}
