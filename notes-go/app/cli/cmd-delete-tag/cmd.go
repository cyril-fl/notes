package cmddeletetag

import (
	"fmt"

	"github.com/spf13/cobra"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/deletetag"
)

func Delete(cobra *cobra.Command, args []string) error {
	process := deletetag.Process(cobra.Context(), postgrey.GetDb(), args)
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully deleted %d tag(s)\n", len(process.GetResult()))
	}

	return process.GetErr()
}
