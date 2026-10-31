package cmdedittag

import (
	"fmt"

	"github.com/spf13/cobra"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/edittag"
)

func Edit(cobra *cobra.Command, args []string) error {
	process := edittag.Process(cobra.Context(), postgrey.GetDb(), args[0], args[1])
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully renamed %q to %q\n", args[0], process.GetResult().Value)
	}

	return process.GetErr()
}
