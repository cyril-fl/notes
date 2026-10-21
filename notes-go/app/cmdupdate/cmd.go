package cmdupdate

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Update(cobra *cobra.Command, args []string) error {
	fmt.Println("Update command executed")
	return nil
}
