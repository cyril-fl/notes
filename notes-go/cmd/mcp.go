package cmd

import (
	cmdmcp "github.com/cyril-fl/notes/notes-go/app/cli/cmd-mcp"
	"github.com/spf13/cobra"
)

var cmdMCP = &cobra.Command{
	Use:     cmd.MCP.Use,
	Short:   cmd.MCP.Short,
	Long:    cmd.MCP.Long,
	Example: cmd.MCP.Example,
	Args:    cobra.NoArgs,
	RunE:    cmdmcp.Serve,
}

func init() {
	cmdRoot.AddCommand(cmdMCP)
}
