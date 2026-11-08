package cmddeletetag

import (
	"fmt"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/deletetag"
	"github.com/spf13/cobra"
)

func Delete(cobra *cobra.Command, args []string) error {
	process := deletetag.Process(cobra.Context(), postgrey.GetDb(), args)
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully deleted %d tag(s)\n", len(process.GetResult()))
	}

	return process.GetErr()
}
