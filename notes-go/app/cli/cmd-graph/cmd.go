package cmdgraph

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	cmdeditnote "main.go/app/cli/cmd-edit-note"
	cmdreadnote "main.go/app/cli/cmd-read-note"
	"main.go/app/shared/services/navigator"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/counttags"
	"main.go/app/shared/services/process/readneighbors"
	"main.go/app/shared/services/process/readtag"
)

// Target kinds of the graph pages besides notes.
const (
	kindTags = "tags"
	kindTag  = "tag"
)

// Graph walks the notes through the tags they share, from a note (or from the tag list).
func Graph(cobra *cobra.Command, args []string) error {
	ctx := cobra.Context()
	start := navigator.Target{Kind: kindTags}
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

// loadPage builds what a node shows: the tag list, the notes of a tag, or the neighbors of a note.
func loadPage(ctx context.Context, target navigator.Target) (navigator.Page, error) {
	switch target.Kind {
	case kindTag:
		return tagPage(ctx, target.Key)
	case navigator.KindNote:
		return notePage(ctx, target.Key)
	default:
		return tagsPage(ctx)
	}
}

func tagsPage(ctx context.Context) (navigator.Page, error) {
	process := counttags.Process(ctx, postgrey.GetDb())
	process.Run()

	tags := process.GetResult()
	p := navigator.Page{Title: "Graph", Path: "tags", Info: count(len(tags), "tag"), Empty: "No tag yet"}
	for _, tag := range tags {
		p.Rows = append(p.Rows, navigator.Row{Cells: []string{tag.Value, count(tag.Notes, "note")}, Target: navigator.Target{Kind: kindTag, Key: tag.Value}})
	}
	return p, process.GetErr()
}

func tagPage(ctx context.Context, value string) (navigator.Page, error) {
	process := readtag.Process(ctx, postgrey.GetDb(), value)
	process.Run()

	result := process.GetResult()
	p := navigator.Page{Title: result.Tag.Value, Path: "tag", Info: count(len(result.Notes), "note"), Empty: "No note carries this tag"}
	for _, note := range result.Notes {
		p.Rows = append(p.Rows, navigator.Row{Cells: []string{untitled(note.Title), note.Path}, Target: navigator.Target{Kind: navigator.KindNote, Key: note.ID}})
	}
	return p, process.GetErr()
}

func notePage(ctx context.Context, id string) (navigator.Page, error) {
	process := readneighbors.Process(ctx, postgrey.GetDb(), id)
	process.Run()

	result := process.GetResult()
	info := []string{count(len(result.Neighbors), "linked note")}
	if len(result.Note.Tags) > 0 {
		info = append([]string{strings.Join(result.Note.Tags, " · ")}, info...)
	}
	p := navigator.Page{
		Title: untitled(result.Note.Item.Title),
		Path:  result.Note.Path,
		Info:  strings.Join(info, " — "),
		Empty: "No note shares a tag with this one",
	}
	for _, neighbor := range result.Neighbors {
		p.Rows = append(p.Rows, navigator.Row{
			Cells:  []string{untitled(neighbor.Title), neighbor.Path, count(neighbor.Shared, "tag"), strings.Join(neighbor.Tags, " · ")},
			Target: navigator.Target{Kind: navigator.KindNote, Key: neighbor.ID},
		})
	}
	return p, process.GetErr()
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
