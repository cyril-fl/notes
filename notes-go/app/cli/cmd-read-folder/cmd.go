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
	var output strings.Builder
	output.WriteString(lipgloss.NewStyle().Bold(true).Render(folder.Item.Title) + "\n")
	output.WriteString(lipgloss.NewStyle().Faint(true).Render(folder.Path) + "\n")

	output.WriteString("\n")
	if len(folder.Children) == 0 {
		output.WriteString(lipgloss.NewStyle().Faint(true).Italic(true).Render("This folder is empty"))
	} else {
		table := tabwriter.NewWriter(&output, 0, 0, 2, ' ', 0)
		for _, child := range folder.Children {
			title := child.Title
			if title == "" {
				title = "untitled"
			}
			fmt.Fprintf(table, "%s\t%s\t%s\n", child.Kind, title, child.ID)
		}
		if err := table.Flush(); err != nil {
			return "", err
		}
	}

	return output.String(), nil
}
