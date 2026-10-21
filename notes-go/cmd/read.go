package cmd

import (
	"github.com/spf13/cobra"
	src "main.go/app/cmdread"
)

var cmdRead = &cobra.Command{
	Use:     cmd.Read.Use,
	Short:   cmd.Read.Short,
	Long:    cmd.Read.Long,
	Example: cmd.Read.Example,
	RunE:    src.Read,
}

func init() {
	cmdRoot.AddCommand(cmdRead)
}
