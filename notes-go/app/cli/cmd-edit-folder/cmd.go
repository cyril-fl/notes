package cmdeditfolder

import (
	"fmt"

	"github.com/spf13/cobra"
	millflagseditfolder "main.go/app/cli/cmd-edit-folder/services/mill"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/editfolder"
)

// Edit loads the folder, applies --title (or a pre-filled form), then saves it.
func Edit(cobra *cobra.Command, args []string) error {
	resolver := millflagseditfolder.New(millflagseditfolder.Template{Ctx: cobra.Context(), Db: postgrey.GetDb(), Flags: cobra.Flags(), ID: args[0]})
	input := resolver.Resolve()
	if resolver.HasErr() {
		return resolver.GetErr()
	}

	process := editfolder.Process(cobra.Context(), postgrey.GetDb(), input)
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully updated folder %s\n", process.GetResult().ID)
	}

	return process.GetErr()
}
