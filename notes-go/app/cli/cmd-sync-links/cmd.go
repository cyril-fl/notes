package cmdsynclinks

import (
	"fmt"

	"github.com/spf13/cobra"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/synclinks"
)

// Sync links every note to the notes its [[links]] name.
func Sync(cobra *cobra.Command, args []string) error {
	process := synclinks.Process(cobra.Context(), postgrey.GetDb())
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully synced %d note(s)\n", process.GetResult())
	}

	return process.GetErr()
}
