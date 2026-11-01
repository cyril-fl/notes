package cmdgraph

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	cmdeditnote "main.go/app/cli/cmd-edit-note"
	cmdreadnote "main.go/app/cli/cmd-read-note"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/counttags"
	"main.go/app/shared/services/process/readneighbors"
	"main.go/app/shared/services/process/readtag"
	"main.go/app/shared/services/screen"
)

// Graph walks the notes through the tags they share, from a note (or from the tag list).
// The note view and the edit form run between two graph screens, which then reopens where it was.
func Graph(cobra *cobra.Command, args []string) error {
	ctx := cobra.Context()
	current := node{kind: nodeTags}
	if len(args) == 1 {
		current = node{kind: nodeNote, key: args[0]}
	}
	load := func(n node) (page, error) { return loadPage(ctx, n) }

	p, err := load(current)
	if err != nil {
		return err
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		screen.Print(render(p, p.rows, -1, p.empty))
		return nil
	}

	var history []node
	cursor := 0
	for {
		final, err := tea.NewProgram(newModel(load, current, p, history, cursor)).Run()
		if err != nil {
			return err
		}
		m := final.(*model)
		if m.err != nil || m.action == actionQuit {
			return m.err
		}

		switch m.action {
		case actionOpen:
			err = cmdreadnote.View(ctx, m.target.key)
		case actionEdit:
			err = cmdeditnote.Interactive(ctx, m.target.key)
		}
		// An aborted form comes back to the graph unchanged; ctrl+q in the note view quits everything.
		if err != nil && !errors.Is(err, huh.ErrUserAborted) {
			return err
		}

		current, history, cursor = m.node, m.history, m.cursor
		if p, err = load(current); err != nil {
			return err
		}
	}
}

// loadPage builds what a node shows: the tag list, the notes of a tag, or the neighbors of a note.
func loadPage(ctx context.Context, n node) (page, error) {
	switch n.kind {
	case nodeTag:
		return tagPage(ctx, n.key)
	case nodeNote:
		return notePage(ctx, n.key)
	default:
		return tagsPage(ctx)
	}
}

func tagsPage(ctx context.Context) (page, error) {
	process := counttags.Process(ctx, postgrey.GetDb())
	process.Run()

	tags := process.GetResult()
	p := page{title: "Graph", path: "tags", info: count(len(tags), "tag"), empty: "No tag yet"}
	for _, tag := range tags {
		p.rows = append(p.rows, row{cells: []string{tag.Value, count(tag.Notes, "note")}, target: node{kind: nodeTag, key: tag.Value}})
	}
	return p, process.GetErr()
}

func tagPage(ctx context.Context, value string) (page, error) {
	process := readtag.Process(ctx, postgrey.GetDb(), value)
	process.Run()

	result := process.GetResult()
	p := page{title: result.Tag.Value, path: "tag", info: count(len(result.Notes), "note"), empty: "No note carries this tag"}
	for _, note := range result.Notes {
		p.rows = append(p.rows, row{cells: []string{untitled(note.Title), note.Path}, target: node{kind: nodeNote, key: note.ID}})
	}
	return p, process.GetErr()
}

func notePage(ctx context.Context, id string) (page, error) {
	process := readneighbors.Process(ctx, postgrey.GetDb(), id)
	process.Run()

	result := process.GetResult()
	info := []string{count(len(result.Neighbors), "linked note")}
	if len(result.Note.Tags) > 0 {
		info = append([]string{strings.Join(result.Note.Tags, " · ")}, info...)
	}
	p := page{
		title: untitled(result.Note.Item.Title),
		path:  result.Note.Path,
		info:  strings.Join(info, " — "),
		empty: "No note shares a tag with this one",
	}
	for _, neighbor := range result.Neighbors {
		p.rows = append(p.rows, row{
			cells:  []string{untitled(neighbor.Title), neighbor.Path, count(neighbor.Shared, "tag"), strings.Join(neighbor.Tags, " · ")},
			target: node{kind: nodeNote, key: neighbor.ID},
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
