package cmddeletenote

import (
	"fmt"

	"github.com/spf13/cobra"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/deleteitems"
)

// Delete moves notes to the trash, by uuid.
func Delete(cobra *cobra.Command, args []string) error {
	process := deleteitems.Process(cobra.Context(), postgrey.GetDb(), domain.NOTE, args)
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully deleted %d note(s)\n", len(process.GetResult()))
	}

	return process.GetErr()
}
