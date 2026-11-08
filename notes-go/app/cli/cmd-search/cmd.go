package cmdsearch

import (
	"context"
	"fmt"
	"strings"

	cmdeditnote "github.com/cyril-fl/notes/notes-go/app/cli/cmd-edit-note"
	cmdreadnote "github.com/cyril-fl/notes/notes-go/app/cli/cmd-read-note"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/navigator"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/searchnotes"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/screen"
	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
)

// Search lists the notes containing the text in their title or content; ↳ opens one, ^E edits it, --json prints them for scripts.
func Search(cobra *cobra.Command, args []string) error {
	ctx := cobra.Context()
	text := strings.Join(args, " ")

	var asJSON bool
	if err := plumcliutilsargs.GetBoolFlags(cobra, map[string]*bool{"json": &asJSON}); err != nil {
		return err
	}
	if asJSON {
		process := searchnotes.Process(ctx, postgrey.GetDb(), text)
		process.Run()
		if !process.HasErr() {
			screen.PrintJSON(process.GetResult())
		}
		return process.GetErr()
	}

	load := func(navigator.Target) (navigator.Page, error) { return searchPage(ctx, text) }
	return navigator.Run(load, navigator.Target{}, navigator.Options{EnterOpens: true}, func(action navigator.Action, note navigator.Target) error {
		if action == navigator.ActionEdit {
			return cmdeditnote.Interactive(ctx, note.Key)
		}
		return cmdreadnote.View(ctx, note.Key)
	})
}

// searchPage is reloaded after an edit: the note may not match anymore.
func searchPage(ctx context.Context, text string) (navigator.Page, error) {
	process := searchnotes.Process(ctx, postgrey.GetDb(), text)
	process.Run()

	matches := process.GetResult()
	info := fmt.Sprintf("%d notes", len(matches))
	if len(matches) == 1 {
		info = "1 note"
	}
	page := navigator.Page{Title: "Search", Path: fmt.Sprintf("%q", text), Info: info, Empty: "No note matches"}
	for _, match := range matches {
		title := match.Title
		if title == "" {
			title = "untitled"
		}
		page.Rows = append(page.Rows, navigator.Row{
			Cells:  []string{title, match.Path, match.Snippet},
			Target: navigator.Target{Kind: navigator.KindNote, Key: match.ID},
		})
	}
	return page, process.GetErr()
}
