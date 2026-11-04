package cmd

import (
	"github.com/spf13/cobra"
	cmdsynctags "main.go/app/cli/cmd-sync-tags"
)

var cmdSyncTags = &cobra.Command{
	Use:     cmd.SyncTags.Use,
	Short:   cmd.SyncTags.Short,
	Long:    cmd.SyncTags.Long,
	Example: cmd.SyncTags.Example,
	Args:    cobra.NoArgs,
	RunE:    cmdsynctags.Sync,
}

func init() {
	cmdRoot.AddCommand(cmdSyncTags)
}
