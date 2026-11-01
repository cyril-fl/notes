package cmdmove

import (
	"fmt"

	"github.com/spf13/cobra"
	millflagsmove "main.go/app/cli/cmd-move/services/mill"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/moveitems"
)

// Move puts notes and folders into the --parent folder, or at the root without it.
func Move(cobra *cobra.Command, args []string) error {
	resolver := millflagsmove.New(millflagsmove.Template{Ctx: cobra.Context(), Db: postgrey.GetDb(), Flags: cobra.Flags(), IDs: args})
	input := resolver.Resolve()
	if resolver.HasErr() {
		return resolver.GetErr()
	}

	process := moveitems.Process(cobra.Context(), postgrey.GetDb(), input)
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully moved %d item(s)\n", len(process.GetResult()))
	}

	return process.GetErr()
}
