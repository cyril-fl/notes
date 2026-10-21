package cmd

import (
	"github.com/spf13/cobra"
	src "main.go/app/cmdgraph"
)

var cmdGraph = &cobra.Command{
	Use:     cmd.Graph.Use,
	Short:   cmd.Graph.Short,
	Long:    cmd.Graph.Long,
	Example: cmd.Graph.Example,
	RunE:    src.Graph,
}

func init() {
	cmdRoot.AddCommand(cmdGraph)
}
