package cmdrecent

import (
	"context"
	"fmt"

	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
	cmdeditnote "main.go/app/cli/cmd-edit-note"
	cmdreadnote "main.go/app/cli/cmd-read-note"
	"main.go/app/shared/services/navigator"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/listrecent"
	"main.go/app/shared/services/screen"
)

// Recent lists the last changed notes; ↳ opens one, ^E edits it, --json prints them for scripts.
func Recent(cobra *cobra.Command, args []string) error {
	ctx := cobra.Context()

	var asJSON bool
	if err := plumcliutilsargs.GetBoolFlags(cobra, map[string]*bool{"json": &asJSON}); err != nil {
		return err
	}
	if asJSON {
		process := listrecent.Process(ctx, postgrey.GetDb())
		process.Run()
		if !process.HasErr() {
			screen.PrintJSON(process.GetResult())
		}
		return process.GetErr()
	}

	load := func(navigator.Target) (navigator.Page, error) { return recentPage(ctx) }
	return navigator.Run(load, navigator.Target{}, navigator.Options{EnterOpens: true}, func(action navigator.Action, note navigator.Target) error {
		if action == navigator.ActionEdit {
			return cmdeditnote.Interactive(ctx, note.Key)
		}
		return cmdreadnote.View(ctx, note.Key)
	})
}

// recentPage is reloaded after an edit: the edited note moves to the top.
func recentPage(ctx context.Context) (navigator.Page, error) {
	process := listrecent.Process(ctx, postgrey.GetDb())
	process.Run()

	notes := process.GetResult()
	info := fmt.Sprintf("%d notes", len(notes))
	if len(notes) == 1 {
		info = "1 note"
	}
	page := navigator.Page{Title: "Recent", Path: "last changed first", Info: info, Empty: "No note yet"}
	for _, note := range notes {
		title := note.Title
		if title == "" {
			title = "untitled"
		}
		page.Rows = append(page.Rows, navigator.Row{
			Cells:  []string{title, note.Path, note.UpdatedAt.Format("2 Jan 2006")},
			Target: navigator.Target{Kind: navigator.KindNote, Key: note.ID},
		})
	}
	return page, process.GetErr()
}
