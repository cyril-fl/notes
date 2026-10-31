package cmdcreatenote

import (
	"fmt"

	"github.com/spf13/cobra"
	millflagscreatenote "main.go/app/cli/cmd-create-note/services/mill"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createnote"
)

// Create resolves the note from flags (and a form for what is missing), then saves it.
func Create(cobra *cobra.Command, args []string) error {
	resolver := millflagscreatenote.New(millflagscreatenote.Template{Ctx: cobra.Context(), Db: postgrey.GetDb(), Flags: cobra.Flags()})
	input := resolver.Resolve()
	if resolver.HasErr() {
		return resolver.GetErr()
	}

	process := createnote.Process(cobra.Context(), postgrey.GetDb(), input)
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully created note %s\n", process.GetResult().ID)
	}

	return process.GetErr()
}
