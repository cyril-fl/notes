package cmdreadnote

import (
	"os"

	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/readitem"
	"main.go/app/shared/services/screen"
)

// Read prints the note path, then its content rendered as markdown.
func Read(cobra *cobra.Command, args []string) error {
	process := readitem.Process(cobra.Context(), postgrey.GetDb(), domain.NOTE, args[0])
	process.Run()
	if process.HasErr() {
		return process.GetErr()
	}

	note := process.GetResult()
	content, err := render(note.Content)
	if err != nil {
		return err
	}

	screen.Print(lipgloss.NewStyle().Faint(true).Render(note.Path) + "\n\n" + content)
	return nil
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
