package cmd

import (
	"github.com/spf13/cobra"
	src "main.go/app/cmdcreate"
)

var cmdCreate = &cobra.Command{
	Use:     cmd.Create.Use,
	Short:   cmd.Create.Short,
	Long:    cmd.Create.Long,
	Example: cmd.Create.Example,
	RunE:    src.Create,
}

func init() {
	cmdRoot.AddCommand(cmdCreate)
}
