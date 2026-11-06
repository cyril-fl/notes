package cmd

import (
	"github.com/spf13/cobra"
	cmdsynctags "main.go/app/cli/cmd-sync-tags"
	cmdsynctitles "main.go/app/cli/cmd-sync-titles"
)

var cmdSyncTags = &cobra.Command{
	Use:     cmd.SyncTags.Use,
	Short:   cmd.SyncTags.Short,
	Long:    cmd.SyncTags.Long,
	Example: cmd.SyncTags.Example,
	Args:    cobra.NoArgs,
	RunE:    cmdsynctags.Sync,
}

var cmdSyncTitles = &cobra.Command{
	Use:     cmd.SyncTitles.Use,
	Short:   cmd.SyncTitles.Short,
	Long:    cmd.SyncTitles.Long,
	Example: cmd.SyncTitles.Example,
	Args:    cobra.NoArgs,
	RunE:    cmdsynctitles.Sync,
}

func init() {
	cmdRoot.AddCommand(cmdSyncTags, cmdSyncTitles)
}
