package cmdeditnote

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Edit(cobra *cobra.Command, args []string) error {
	fmt.Println("Update command executed")
	return nil
}
