package cmdreadnote

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Read(cobra *cobra.Command, args []string) error {
	fmt.Println("Finder command executed")
	return nil
}
