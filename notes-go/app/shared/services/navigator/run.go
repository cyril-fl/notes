package navigator

import (
	"errors"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"golang.org/x/term"
	"main.go/app/shared/services/screen"
)

// Act opens or edits the note of a row; the screen reopens where it was afterwards.
type Act func(action Action, note Target) error

// Run shows the start page and lets the user walk from page to page until they quit.
// Opening or editing a note quits the screen, runs act, then reopens it (history and cursor kept).
// A pipe or file gets the start page once.
func Run(load Loader, start Target, options Options, act Act) error {
	page, err := load(start)
	if err != nil {
		return err
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		screen.Print(Render(page, page.Rows, -1, page.Empty))
		return nil
	}

	current, cursor := start, 0
	var history []Target
	for {
		final, err := tea.NewProgram(newModel(load, options, current, page, history, cursor)).Run()
		if err != nil {
			return err
		}
		m := final.(*model)
		if m.err != nil || m.action == ActionQuit {
			return m.err
		}

		// An aborted form comes back to the screen unchanged; ctrl+q in the note view quits everything.
		if err := act(m.action, m.target); err != nil && !errors.Is(err, huh.ErrUserAborted) {
			return err
		}

		current, history, cursor = m.current, m.history, m.cursor
		if page, err = load(current); err != nil {
			return err
		}
	}
}
