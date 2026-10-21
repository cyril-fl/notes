package cmd

import (
	"github.com/spf13/cobra"
	src "main.go/app/cmdupdate"
)

var cmdUpdate = &cobra.Command{
	Use:     cmd.Update.Use,
	Short:   cmd.Update.Short,
	Long:    cmd.Update.Long,
	Example: cmd.Update.Example,
	RunE:    src.Update,
}

func init() {
	cmdRoot.AddCommand(cmdUpdate)
}
