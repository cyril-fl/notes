package cmdsynctitles

import (
	"fmt"

	"github.com/spf13/cobra"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/synctitles"
)

// Sync makes the title of every note come from its text.
func Sync(cobra *cobra.Command, args []string) error {
	process := synctitles.Process(cobra.Context(), postgrey.GetDb())
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully synced %d note(s)\n", process.GetResult())
	}

	return process.GetErr()
}
