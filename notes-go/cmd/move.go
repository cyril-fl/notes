package cmd

import (
	"github.com/spf13/cobra"
	cmdmove "main.go/app/cli/cmd-move"
)

var cmdMove = &cobra.Command{
	Use:     cmd.Move.Use,
	Short:   cmd.Move.Short,
	Long:    cmd.Move.Long,
	Example: cmd.Move.Example,
	RunE:    cmdmove.Move,
}

func init() {
	cmdRoot.AddCommand(cmdMove)
}
