package cmdrestore

import (
	"fmt"

	"github.com/spf13/cobra"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/restoreitems"
)

// Restore takes items out of the trash, by uuid, with everything removed at the same time.
func Restore(cobra *cobra.Command, args []string) error {
	process := restoreitems.Process(cobra.Context(), postgrey.GetDb(), args)
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully restored %d item(s)\n", len(process.GetResult()))
	}

	return process.GetErr()
}
