package cmd

import (
	"github.com/spf13/cobra"
	cmdgraph "main.go/app/cli/cmd-graph"
)

var cmdGraph = &cobra.Command{
	Use:     cmd.ShowGraph.Use,
	Short:   cmd.ShowGraph.Short,
	Long:    cmd.ShowGraph.Long,
	Example: cmd.ShowGraph.Example,
	Args:    cobra.MaximumNArgs(1),
	RunE:    cmdgraph.Graph,
}

func init() {
	cmdRoot.AddCommand(cmdGraph)
}
