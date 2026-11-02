// Package screen prints command output the way the forms show: as tall as the terminal.
package screen

import (
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	plumdebug "github.com/cyril-fl/plum/utils/debug"
	"golang.org/x/term"
)

// Print writes output, padded with blank lines to the terminal height; a pipe or file gets it as is.
func Print(output string) {
	output = strings.TrimRight(output, "\n")
	if _, height, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		output = pad(output, height)
	}
	lipgloss.Println(output)
}

// PrintJSON writes a list for scripts: an empty one is [], not null.
func PrintJSON[T any](items []T) {
	plumdebug.Json(append([]T{}, items...))
}

// pad keeps the last line for the shell prompt; longer output is left as is.
func pad(output string, height int) string {
	return output + strings.Repeat("\n", max(0, height-1-lipgloss.Height(output)))
}
