package cmdreadfolder

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"charm.land/huh/v2"
	cmdeditfolder "github.com/cyril-fl/notes/notes-go/app/cli/cmd-edit-folder"
	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/screen"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/style"
	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	plumdebug "github.com/cyril-fl/plum/utils/debug"
	"github.com/spf13/cobra"
)

// Read shows the folder title and path, then one line per direct child with its uuid; ctrl+e edits it; --json prints it for scripts.
func Read(cobra *cobra.Command, args []string) error {
	var asJSON bool
	if err := plumcliutilsargs.GetBoolFlags(cobra, map[string]*bool{"json": &asJSON}); err != nil {
		return err
	}
	if asJSON {
		return printJSON(cobra.Context(), args[0])
	}

	for {
		output, err := load(cobra.Context(), args[0])
		if err != nil {
			return err
		}

		edit, err := screen.Show(output)
		if err != nil || !edit {
			return err
		}
		// An aborted form goes back to the view unchanged.
		if err := cmdeditfolder.Interactive(cobra.Context(), args[0]); err != nil && !errors.Is(err, huh.ErrUserAborted) {
			return err
		}
	}
}

// printJSON gives an empty folder "children": [], not null.
func printJSON(ctx context.Context, id string) error {
	process := readitem.Process(ctx, postgrey.GetDb(), domain.FOLDER, id)
	process.Run()
	if process.HasErr() {
		return process.GetErr()
	}

	folder := process.GetResult()
	folder.Children = append([]domain.Item{}, folder.Children...)
	plumdebug.Json(folder)
	return nil
}

func load(ctx context.Context, id string) (string, error) {
	process := readitem.Process(ctx, postgrey.GetDb(), domain.FOLDER, id)
	process.Run()
	if process.HasErr() {
		return "", process.GetErr()
	}

	folder := process.GetResult()
	return Output(folder, folder.Children, -1, EmptyFolder), nil
}

// EmptyFolder is the message of a folder without children.
const EmptyFolder = "This folder is empty"

// Output renders a folder like `show folder`: the header from the whole folder, then rows (the finder
// passes its filtered, scrolled window); cursor >= 0 marks that row as selected, empty shows when no row.
func Output(folder readitem.Result, rows []domain.Item, cursor int, empty string) string {
	header := style.Header(string(folder.Item.Title), folder.Path, info(folder)) + "\n"
	if len(rows) == 0 {
		return header + style.Empty().Render(empty)
	}

	cells := make([][]string, len(rows))
	for i, child := range rows {
		title := child.Title
		if title == "" {
			title = "untitled"
		}
		cells[i] = []string{string(child.Kind), string(title), child.ID}
	}
	return header + style.Table(cells, cursor, 1)
}

// info counts the children, and dates the folder when it is a real one (the root has no row).
func info(folder readitem.Result) string {
	folders := 0
	for _, child := range folder.Children {
		if child.Kind == domain.FOLDER {
			folders++
		}
	}
	parts := []string{count(folders, "folder"), count(len(folder.Children)-folders, "note")}

	updated := folder.Item.UpdatedAt
	if updated.IsZero() {
		updated = folder.Item.CreatedAt
	}
	if !updated.IsZero() {
		parts = append(parts, "updated "+updated.Format("2 Jan 2006"))
	}
	return strings.Join(parts, " · ")
}

func count(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}
