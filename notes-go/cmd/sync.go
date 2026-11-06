package cmd

import (
	"github.com/spf13/cobra"
	cmdsynclinks "main.go/app/cli/cmd-sync-links"
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

var cmdSyncLinks = &cobra.Command{
	Use:     cmd.SyncLinks.Use,
	Short:   cmd.SyncLinks.Short,
	Long:    cmd.SyncLinks.Long,
	Example: cmd.SyncLinks.Example,
	Args:    cobra.NoArgs,
	RunE:    cmdsynclinks.Sync,
}

func init() {
	cmdRoot.AddCommand(cmdSyncTags, cmdSyncTitles, cmdSyncLinks)
}
