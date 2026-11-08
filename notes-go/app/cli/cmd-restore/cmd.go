package cmdrestore

import (
	"fmt"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/restoreitems"
	"github.com/spf13/cobra"
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
