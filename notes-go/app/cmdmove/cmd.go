package cmdmove

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Move(cobra *cobra.Command, args []string) error {
	fmt.Println("Move command executed")
	return nil
}
