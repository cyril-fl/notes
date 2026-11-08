package cmdcreatetag

import (
	"fmt"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createtag"
	"github.com/spf13/cobra"
)

func Create(cobra *cobra.Command, args []string) error {
	process := createtag.Process(cobra.Context(), postgrey.GetDb(), args)
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully created %d tag(s)\n", len(process.GetResult()))
	}

	return process.GetErr()
}
