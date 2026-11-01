package cmd

import (
	"github.com/spf13/cobra"
	cmdrestore "main.go/app/cli/cmd-restore"
	cmdtrash "main.go/app/cli/cmd-trash"
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
	cmdRoot.AddCommand(cmdTrash, cmdRestore)
}
