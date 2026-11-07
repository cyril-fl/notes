package cmdsyncentities

import (
	"fmt"

	"github.com/spf13/cobra"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/syncentities"
)

// Sync repairs the notes whose text holds & < > as HTML entities.
func Sync(cobra *cobra.Command, args []string) error {
	process := syncentities.Process(cobra.Context(), postgrey.GetDb())
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully synced %d note(s)\n", process.GetResult())
	}

	return process.GetErr()
}
