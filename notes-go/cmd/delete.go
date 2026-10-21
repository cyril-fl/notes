package cmd

import (
	"github.com/spf13/cobra"
	src "main.go/app/cmddelete"
)

var cmdDelete = &cobra.Command{
	Use:     cmd.Delete.Use,
	Short:   cmd.Delete.Short,
	Long:    cmd.Delete.Long,
	Example: cmd.Delete.Example,
	RunE:    src.Delete,
}

func init() {
	cmdRoot.AddCommand(cmdDelete)
}
