package cmdreadfolder

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/readitem"
	"main.go/app/shared/services/screen"
)

// Read prints the folder title and path, then one line per direct child with its uuid.
func Read(cobra *cobra.Command, args []string) error {
	process := readitem.Process(cobra.Context(), postgrey.GetDb(), domain.FOLDER, args[0])
	process.Run()
	if process.HasErr() {
		return process.GetErr()
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
			return err
		}
	}

	screen.Print(output.String())
	return nil
}
