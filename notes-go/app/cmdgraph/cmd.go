package cmdgraph

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Graph(cobra *cobra.Command, args []string) error {
	fmt.Println("Graph command executed")
	return nil
}
