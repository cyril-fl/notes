package cmdreadtag

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Read(cobra *cobra.Command, args []string) error {
	fmt.Println("Tag command executed")
	return nil
}
