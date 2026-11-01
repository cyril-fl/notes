package cmdeditnote

import (
	"fmt"

	"github.com/spf13/cobra"
	millflagseditnote "main.go/app/cli/cmd-edit-note/services/mill"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/editnote"
)

// Edit loads the note, applies the flags (or a pre-filled form when none), then saves it.
func Edit(cobra *cobra.Command, args []string) error {
	resolver := millflagseditnote.New(millflagseditnote.Template{Ctx: cobra.Context(), Db: postgrey.GetDb(), Flags: cobra.Flags(), ID: args[0]})
	input := resolver.Resolve()
	if resolver.HasErr() {
		return resolver.GetErr()
	}

	process := editnote.Process(cobra.Context(), postgrey.GetDb(), input)
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully updated note %s\n", process.GetResult().ID)
	}

	return process.GetErr()
}
