package cmdgraph

import (
	"context"
	"fmt"
	"strings"

	cmdeditnote "github.com/cyril-fl/notes/notes-go/app/cli/cmd-edit-note"
	cmdreadnote "github.com/cyril-fl/notes/notes-go/app/cli/cmd-read-note"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/navigator"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readgraph"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readneighbors"
	"github.com/spf13/cobra"
)

// kindNotes is the start page: every note.
const kindNotes = "notes"

// Graph walks the notes through their [[links]] and the tags they share, from a note (or from every note).
func Graph(cobra *cobra.Command, args []string) error {
	ctx := cobra.Context()
	start := navigator.Target{Kind: kindNotes}
	if len(args) == 1 {
		start = navigator.Target{Kind: navigator.KindNote, Key: args[0]}
	}

	load := func(target navigator.Target) (navigator.Page, error) { return loadPage(ctx, target) }
	return navigator.Run(load, start, navigator.Options{}, func(action navigator.Action, note navigator.Target) error {
		if action == navigator.ActionEdit {
			return cmdeditnote.Interactive(ctx, note.Key)
		}
		return cmdreadnote.View(ctx, note.Key)
	})
}

// loadPage builds what a node shows: every note, or the neighbors of a note.
func loadPage(ctx context.Context, target navigator.Target) (navigator.Page, error) {
	if target.Kind == navigator.KindNote {
		return notePage(ctx, target.Key)
	}
	return notesPage(ctx)
}

// notesPage lists every note with how many notes it is linked to, like the web graph.
func notesPage(ctx context.Context) (navigator.Page, error) {
	process := readgraph.Process(ctx, postgrey.GetDb())
	process.Run()

	result := process.GetResult()
	linked := map[string]int{}
	for _, link := range result.Links {
		linked[link.Source]++
		linked[link.Target]++
	}
	p := navigator.Page{Title: "Graph", Path: "notes", Info: count(len(result.Notes), "note"), Empty: "No note yet"}
	for _, note := range result.Notes {
		p.Rows = append(p.Rows, navigator.Row{
			Cells:  []string{untitled(string(note.Title)), note.Path, count(linked[note.ID], "linked note")},
			Target: navigator.Target{Kind: navigator.KindNote, Key: note.ID},
		})
	}
	return p, process.GetErr()
}

func notePage(ctx context.Context, id string) (navigator.Page, error) {
	process := readneighbors.Process(ctx, postgrey.GetDb(), id)
	process.Run()

	result := process.GetResult()
	rows := append(linkedRows(result.Links, "link"), linkedRows(result.Backlinks, "backlink")...)
	info := []string{count(len(rows)+len(result.Neighbors), "linked note")}
	if len(result.Note.Tags) > 0 {
		info = append([]string{strings.Join(result.Note.Tags, " · ")}, info...)
	}
	p := navigator.Page{
		Title: untitled(string(result.Note.Item.Title)),
		Path:  result.Note.Path,
		Info:  strings.Join(info, " — "),
		Empty: "No note links to or shares a tag with this one",
		Rows:  rows,
	}
	for _, neighbor := range result.Neighbors {
		p.Rows = append(p.Rows, navigator.Row{
			Cells:  []string{untitled(string(neighbor.Title)), neighbor.Path, count(neighbor.Shared, "tag"), strings.Join(neighbor.Tags, " · ")},
			Target: navigator.Target{Kind: navigator.KindNote, Key: neighbor.ID},
		})
	}
	return p, process.GetErr()
}

// linkedRows lists the notes at the other end of the [[links]], same columns as the neighbors.
func linkedRows(notes []readneighbors.Linked, relation string) []navigator.Row {
	rows := make([]navigator.Row, 0, len(notes))
	for _, note := range notes {
		rows = append(rows, navigator.Row{
			Cells:  []string{untitled(string(note.Title)), note.Path, relation, ""},
			Target: navigator.Target{Kind: navigator.KindNote, Key: note.ID},
		})
	}
	return rows
}

// Helpers

func untitled(title string) string {
	if title == "" {
		return "untitled"
	}
	return title
}

func count(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}
