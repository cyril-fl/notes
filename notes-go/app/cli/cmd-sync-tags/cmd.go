package cmdsynctags

import (
	"fmt"

	"github.com/spf13/cobra"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/synctags"
)

// Sync makes the tags of every note match the #tags of its text.
func Sync(cobra *cobra.Command, args []string) error {
	process := synctags.Process(cobra.Context(), postgrey.GetDb())
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully synced %d note(s)\n", process.GetResult())
	}

	return process.GetErr()
}
