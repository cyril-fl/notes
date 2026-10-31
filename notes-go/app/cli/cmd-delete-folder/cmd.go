package cmddeletefolder

import (
	"fmt"

	"github.com/spf13/cobra"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/deleteitems"
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
