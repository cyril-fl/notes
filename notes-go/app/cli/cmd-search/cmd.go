package cmdsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	cmdeditnote "main.go/app/cli/cmd-edit-note"
	cmdreadnote "main.go/app/cli/cmd-read-note"
	"main.go/app/shared/services/navigator"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/searchnotes"
)

// Search lists the notes containing the text in their title or content; ↳ opens one, ^E edits it.
func Search(cobra *cobra.Command, args []string) error {
	ctx := cobra.Context()
	text := strings.Join(args, " ")

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
