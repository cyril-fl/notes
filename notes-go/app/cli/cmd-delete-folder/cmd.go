package cmddeletefolder

import (
	"fmt"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/deleteitems"
	"github.com/spf13/cobra"
)

// Delete moves folders and their whole content to the trash, by uuid.
func Delete(cobra *cobra.Command, args []string) error {
	process := deleteitems.Process(cobra.Context(), postgrey.GetDb(), domain.FOLDER, args)
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully deleted %d item(s): folders and their content\n", len(process.GetResult()))
	}

	return process.GetErr()
}
