package cmdfinder

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Finder(cobra *cobra.Command, args []string) error {
	fmt.Println("Finder command executed")
	return nil
}
