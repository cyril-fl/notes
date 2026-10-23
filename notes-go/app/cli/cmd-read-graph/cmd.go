package cmdreadgraph

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Read(cobra *cobra.Command, args []string) error {
	fmt.Println("Graph command executed")
	return nil
}
