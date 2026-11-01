package cmdreadnote

import (
	"context"
	"errors"
	"os"

	"charm.land/glamour/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	cmdeditnote "main.go/app/cli/cmd-edit-note"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/readitem"
	"main.go/app/shared/services/screen"
)

// Read shows the note path, then its content rendered as markdown; ctrl+e edits it, then shows it again.
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
		if err := cmdeditnote.Interactive(cobra.Context(), args[0]); err != nil && !errors.Is(err, huh.ErrUserAborted) {
			return err
		}
	}
}

func load(ctx context.Context, id string) (string, error) {
	process := readitem.Process(ctx, postgrey.GetDb(), domain.NOTE, id)
	process.Run()
	if process.HasErr() {
		return "", process.GetErr()
	}

	note := process.GetResult()
	content, err := render(note.Content)
	if err != nil {
		return "", err
	}

	return lipgloss.NewStyle().Faint(true).Render(note.Path) + "\n\n" + content, nil
}

// render keeps the raw markdown when stdout is not a terminal (pipe, file).
func render(markdown string) (string, error) {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return markdown + "\n", nil
	}

	style := "light"
	if lipgloss.HasDarkBackground(os.Stdin, os.Stdout) {
		style = "dark"
	}

	renderer, err := glamour.NewTermRenderer(glamour.WithStandardStyle(style), glamour.WithWordWrap(width))
	if err != nil {
		return "", err
	}
	return renderer.Render(markdown)
}
