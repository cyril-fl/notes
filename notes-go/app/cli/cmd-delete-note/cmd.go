package cmddeletenote

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Delete(cobra *cobra.Command, args []string) error {
	fmt.Println("Delete command executed")
	return nil
}
