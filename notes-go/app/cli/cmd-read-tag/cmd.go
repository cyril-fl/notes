package cmdreadtag

import (
	"fmt"

	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	plumdebug "github.com/cyril-fl/plum/utils/debug"
	"github.com/spf13/cobra"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/listtag"
)

// Read lists tags with --list, one value per line, or as YAML/JSON with --yaml/--json.
func Read(cobra *cobra.Command, args []string) error {
	var asList, asYAML, asJSON bool
	if err := plumcliutilsargs.GetBoolFlags(cobra, map[string]*bool{"list": &asList, "yaml": &asYAML, "json": &asJSON}); err != nil {
		return err
	}

	// shortcut: only --list exists, showing a single tag comes with notes.
	if !asList {
		return cobra.Help()
	}

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
