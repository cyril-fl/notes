package cmdimport

import (
	"fmt"
	"os"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/importnotes"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/resolveitem"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/silo"
	"github.com/spf13/cobra"
)

// Import adds the .md files of the directory args[0] as notes, and its directories as folders, into --parent.
func Import(cobra *cobra.Command, args []string) error {
	input := importnotes.Input{Dir: os.DirFS(args[0])}
	if parent, _ := cobra.Flags().GetString("parent"); parent != "" {
		// Not cmd's resolved(): it would read the directory, a path too, as an item.
		resolve := resolveitem.Process(cobra.Context(), postgrey.GetDb(), domain.FOLDER, []string{parent})
		resolve.Run()
		if resolve.HasErr() {
			return resolve.GetErr()
		}
		input.ParentID = resolve.GetResult()[0]
	}
	if info, err := os.Stat(args[0]); err != nil || !info.IsDir() {
		return fmt.Errorf("not a directory: %s", args[0])
	}

	store, err := silo.New(cobra.Context())
	if err != nil {
		return err
	}
	process := importnotes.Process(cobra.Context(), postgrey.GetDb(), store, input)
	process.Run()

	if !process.HasErr() {
		result := process.GetResult()
		fmt.Printf("Successfully imported %d note(s) and %d folder(s)\n", result.Notes, result.Folders)
	}

	return process.GetErr()
}
