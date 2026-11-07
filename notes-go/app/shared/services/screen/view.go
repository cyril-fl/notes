package screen

import (
	"errors"
	"os"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	plumtypes "github.com/cyril-fl/plum/types"
	"golang.org/x/term"
	"main.go/app/shared/services/style"
)

// ErrQuit is returned when the user quits everything (ctrl+q), not just this view (esc):
// a caller showing it from another screen stops too; cmd.Execute keeps it silent.
var ErrQuit = errors.New("user quit")

// footerHeight is the blank line and the shortcut line under the content.
const footerHeight = 2

type view struct {
	content  viewport.Model
	edit     bool
	quit     bool
	quitting bool
}

// Show displays output as tall as the terminal, scrollable, until the user leaves it.
// It reports true when the user asked to edit the item (ctrl+e), ErrQuit on ctrl+q; a pipe or file just gets the output.
func Show(output string) (bool, error) {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		Print(output)
		return false, nil
	}

	content := viewport.New()
	content.SetContent(strings.TrimRight(output, "\n"))
	model, err := tea.NewProgram(&view{content: content}).Run()
	if err != nil {
		return false, err
	}
	if model.(*view).quit {
		return false, ErrQuit
	}
	return model.(*view).edit, nil
}

func (v *view) Init() tea.Cmd {
	return nil
}

func (v *view) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// The last terminal line stays for the shell prompt, like Print.
		v.content.SetWidth(msg.Width)
		v.content.SetHeight(max(1, msg.Height-1-footerHeight))
		return v, nil
	case tea.KeyPressMsg:
		switch plumtypes.Key(msg.String()) {
		case plumtypes.SHORT_CTRL_E:
			v.edit, v.quitting = true, true
			return v, tea.Quit
		case plumtypes.KEY_ESC:
			v.quitting = true
			return v, tea.Quit
		case plumtypes.SHORT_CTRL_Q, plumtypes.SHORT_CTRL_C:
			v.quit, v.quitting = true, true
			return v, tea.Quit
		}
	}

	var cmd tea.Cmd
	v.content, cmd = v.content.Update(msg)
	return v, cmd
}

// View clears itself before an edit, so the form takes the screen; on quit the content stays.
func (v *view) View() tea.View {
	if v.edit {
		return tea.NewView("")
	}
	if v.quitting {
		return tea.NewView(v.content.View())
	}
	return tea.NewView(v.content.View() + "\n\n" + footer())
}

// Hint is one shortcut of a footer line.
type Hint struct {
	Key   plumtypes.Key
	Label string
}

// Footer renders shortcuts like the forms' help line: accent key glyph, muted label, subtle " · " between.
func Footer(hints ...Hint) string {
	parts := make([]string, len(hints))
	for i, hint := range hints {
		parts[i] = style.Key().Render(hint.Key.Symbol()) + " " + style.Muted().Render(hint.Label)
	}
	return strings.Join(parts, style.Subtle().Render(" · "))
}

func footer() string {
	return Footer(Hint{plumtypes.SHORT_CTRL_E, "Edit"}, Hint{plumtypes.KEY_ESC, "Back"}, Hint{plumtypes.SHORT_CTRL_Q, "Quit"})
}
