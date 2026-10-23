package cmdcreatetag

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Create(cobra *cobra.Command, args []string) error {
	fmt.Println("Create command executed with args:", args)
	return nil
}
