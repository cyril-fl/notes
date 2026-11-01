package cmdreadtag

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	plumdebug "github.com/cyril-fl/plum/utils/debug"
	"github.com/spf13/cobra"
	cmdedittag "main.go/app/cli/cmd-edit-tag"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/listtag"
	"main.go/app/shared/services/process/readtag"
	"main.go/app/shared/services/screen"
)

// Read shows one tag and its notes, or lists tags with --list, one value per line, or as YAML/JSON with --yaml/--json.
func Read(cobra *cobra.Command, args []string) error {
	var asList, asYAML, asJSON bool
	if err := plumcliutilsargs.GetBoolFlags(cobra, map[string]*bool{"list": &asList, "yaml": &asYAML, "json": &asJSON}); err != nil {
		return err
	}

	switch {
	case asList:
		return listTags(cobra, asYAML, asJSON)
	case len(args) == 1:
		return showTag(cobra, args[0])
	default:
		return cobra.Help()
	}
}

// showTag shows the tag, then the path and uuid of each note carrying it; ctrl+e renames it.
func showTag(cobra *cobra.Command, value string) error {
	for {
		output, err := loadTag(cobra.Context(), value)
		if err != nil {
			return err
		}

		edit, err := screen.Show(output)
		if err != nil || !edit {
			return err
		}
		// An aborted form goes back to the view unchanged; a rename reloads under the new value.
		if value, err = cmdedittag.Interactive(cobra.Context(), value); err != nil && !errors.Is(err, huh.ErrUserAborted) {
			return err
		}
	}
}

func loadTag(ctx context.Context, value string) (string, error) {
	process := readtag.Process(ctx, postgrey.GetDb(), value)
	process.Run()
	if process.HasErr() {
		return "", process.GetErr()
	}

	result := process.GetResult()
	var output strings.Builder
	output.WriteString(lipgloss.NewStyle().Bold(true).Render(result.Tag.Value) + "\n")

	output.WriteString("\n")
	if len(result.Notes) == 0 {
		output.WriteString(lipgloss.NewStyle().Faint(true).Italic(true).Render("No note carries this tag"))
	} else {
		table := tabwriter.NewWriter(&output, 0, 0, 2, ' ', 0)
		for _, note := range result.Notes {
			fmt.Fprintf(table, "%s\t%s\n", note.Path, note.ID)
		}
		if err := table.Flush(); err != nil {
			return "", err
		}
	}

	return output.String(), nil
}

func listTags(cobra *cobra.Command, asYAML bool, asJSON bool) error {
	process := listtag.Process(cobra.Context(), postgrey.GetDb())
	process.Run()

	if !process.HasErr() {
		printTags(process.GetResult(), asYAML, asJSON)
	}

	return process.GetErr()
}

// printTags keeps the plain output to values only: ids are for scripts, through --yaml/--json.
func printTags(tags []domain.Tag, asYAML bool, asJSON bool) {
	switch {
	case asJSON:
		plumdebug.Json(tags)
	case asYAML:
		plumdebug.Yaml(tags)
	default:
		for _, tag := range tags {
			fmt.Println(tag.Value)
		}
	}
}
