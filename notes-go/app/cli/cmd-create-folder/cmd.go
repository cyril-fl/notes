package cmdcreatefolder

import (
	"fmt"

	"github.com/spf13/cobra"
	millflagscreatefolder "main.go/app/cli/cmd-create-folder/services/mill"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createfolder"
)

// Create resolves the folder from flags (and a form for a missing title), then saves it.
func Create(cobra *cobra.Command, args []string) error {
	resolver := millflagscreatefolder.New(millflagscreatefolder.Template{Ctx: cobra.Context(), Db: postgrey.GetDb(), Flags: cobra.Flags()})
	input := resolver.Resolve()
	if resolver.HasErr() {
		return resolver.GetErr()
	}

	process := createfolder.Process(cobra.Context(), postgrey.GetDb(), input)
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully created folder %s\n", process.GetResult().ID)
	}

	return process.GetErr()
}
