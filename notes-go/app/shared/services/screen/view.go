package screen

import (
	"os"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	plumtypes "github.com/cyril-fl/plum/types"
	"golang.org/x/term"
)

// footerHeight is the blank line and the shortcut line under the content.
const footerHeight = 2

type view struct {
	content  viewport.Model
	edit     bool
	quitting bool
}

// Show displays output as tall as the terminal, scrollable, until the user quits.
// It reports true when the user asked to edit the item (ctrl+e); a pipe or file just gets the output.
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
		case plumtypes.KEY_ESC, plumtypes.SHORT_CTRL_C, "q":
			v.quitting = true
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

// footer renders the shortcuts like the forms' help line: bold key, muted label, " · " between.
func footer() string {
	key, label := lipgloss.NewStyle().Bold(true), lipgloss.NewStyle().Faint(true)
	return strings.Join([]string{
		key.Render(plumtypes.SHORT_CTRL_E.Symbol()) + " " + label.Render("Edit"),
		key.Render(plumtypes.KEY_ESC.Symbol()) + " " + label.Render("Quit"),
	}, label.Render(" · "))
}
