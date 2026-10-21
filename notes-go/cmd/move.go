package cmd

import (
	"github.com/spf13/cobra"
	src "main.go/app/cmdmove"
)

var cmdMove = &cobra.Command{
	Use:     cmd.Move.Use,
	Short:   cmd.Move.Short,
	Long:    cmd.Move.Long,
	Example: cmd.Move.Example,
	RunE:    src.Move,
}

func init() {
	cmdRoot.AddCommand(cmdMove)
}
