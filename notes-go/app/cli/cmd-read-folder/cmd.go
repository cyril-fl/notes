package cmdreadfolder

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
	cmdeditfolder "main.go/app/cli/cmd-edit-folder"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/readitem"
	"main.go/app/shared/services/screen"
)

// Read shows the folder title and path, then one line per direct child with its uuid; ctrl+e edits it.
func Read(cobra *cobra.Command, args []string) error {
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
	var output strings.Builder
	output.WriteString(lipgloss.NewStyle().Bold(true).Render(folder.Item.Title) + "\n")
	output.WriteString(lipgloss.NewStyle().Faint(true).Render(folder.Path) + "\n")
	output.WriteString(lipgloss.NewStyle().Faint(true).Render(info(folder)) + "\n")

	output.WriteString("\n")
	if len(rows) == 0 {
		output.WriteString(lipgloss.NewStyle().Faint(true).Italic(true).Render(empty))
	} else {
		var lines strings.Builder
		table := tabwriter.NewWriter(&lines, 0, 0, 2, ' ', 0)
		for _, child := range rows {
			title := child.Title
			if title == "" {
				title = "untitled"
			}
			fmt.Fprintf(table, "%s\t%s\t%s\n", child.Kind, title, child.ID)
		}
		table.Flush() // writes to a strings.Builder: cannot fail

		// Styled after the tabwriter: escape codes would break its column widths.
		for i, row := range strings.Split(strings.TrimRight(lines.String(), "\n"), "\n") {
			switch {
			case cursor < 0:
				output.WriteString(row + "\n")
			case i == cursor:
				output.WriteString(lipgloss.NewStyle().Bold(true).Render("> "+row) + "\n")
			default:
				output.WriteString("  " + row + "\n")
			}
		}
	}

	return output.String()
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
