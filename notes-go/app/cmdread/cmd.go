package cmdread

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Read(cobra *cobra.Command, args []string) error {
	fmt.Println("Read command executed")
	return nil
}
