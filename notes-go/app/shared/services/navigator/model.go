// Package navigator is a navigable list screen shared by graph and search: rows lead to other pages or to notes,
// with history, filter and the same keys as the finder. Opening or editing a note is handed back to the caller.
package navigator

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	plumtypes "github.com/cyril-fl/plum/types"
	"main.go/app/shared/services/screen"
)

// headerHeight is the title, path, info and blank line above the rows; footerHeight the blank and shortcut lines.
const (
	headerHeight = 4
	footerHeight = 2
)

// KindNote marks a target that is a note: open (^O) and edit (^E) apply to it. Other kinds belong to the caller.
const KindNote = "note"

// Target is a place a row leads to: Kind is the caller's (KindNote for a note), Key the value it needs to load it.
type Target struct {
	Kind string
	Key  string
}

// Row is one line of a page; Target is where ↳ goes.
type Row struct {
	Cells  []string
	Target Target
}

// Page is what a target shows, rendered like show folder.
type Page struct {
	Title, Path, Info, Empty string
	Rows                     []Row
}

// Loader builds the page of a target.
type Loader func(Target) (Page, error)

// Options says what ↳ does on a note row: center on it (graph) or open it (search).
type Options struct {
	EnterOpens bool
}

type loadedMsg struct {
	target Target
	page   Page
	back   bool
	err    error
}

// Action is what the screen hands back when it quits: open or edit a note, or nothing.
type Action int

const (
	ActionQuit Action = iota
	ActionOpen
	ActionEdit
)

type model struct {
	load      Loader
	options   Options
	current   Target
	page      Page
	history   []Target
	cursor    int
	offset    int
	height    int
	filter    textinput.Model
	filtering bool
	action    Action
	target    Target
	err       error
}

func newModel(load Loader, options Options, current Target, page Page, history []Target, cursor int) *model {
	m := &model{load: load, options: options, current: current, page: page, history: history, filter: textinput.New()}
	m.filter.Prompt = "/"
	m.cursor = min(max(0, cursor), max(0, len(m.visible())-1))
	return m
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
	case loadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		if msg.back {
			m.history = m.history[:len(m.history)-1]
		} else {
			m.history = append(m.history, m.current)
		}
		m.current, m.page, m.cursor, m.offset = msg.target, msg.page, 0, 0
		m.filter.SetValue("")
	case tea.KeyPressMsg:
		if m.filtering {
			return m.updateFilter(msg)
		}
		return m.updateKeys(msg)
	}
	return m, nil
}

func (m *model) updateFilter(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch plumtypes.Key(msg.String()) {
	case plumtypes.KEY_ESC:
		m.filter.SetValue("")
		fallthrough
	case plumtypes.KEY_ENTER:
		m.filtering = false
		m.filter.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	m.cursor, m.offset = 0, 0
	return m, cmd
}

func (m *model) updateKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	selected, hasSelected := m.selected()
	isNote := hasSelected && selected.Target.Kind == KindNote

	switch plumtypes.Key(msg.String()) {
	case plumtypes.KEY_UP, "k":
		m.cursor = max(0, m.cursor-1)
	case plumtypes.KEY_DOWN, "j":
		m.cursor = max(0, min(len(m.visible())-1, m.cursor+1)) // max: an empty list keeps the cursor at 0
	case plumtypes.KEY_ENTER, plumtypes.KEY_RIGHT:
		switch {
		case isNote && m.options.EnterOpens:
			return m.quit(ActionOpen, selected.Target)
		case hasSelected:
			return m, m.open(selected.Target, false)
		}
	case plumtypes.KEY_LEFT, plumtypes.KEY_BACKSPACE:
		return m, m.back()
	case plumtypes.SHORT_CTRL_O:
		if isNote {
			return m.quit(ActionOpen, selected.Target)
		}
	case plumtypes.SHORT_CTRL_E:
		if isNote {
			return m.quit(ActionEdit, selected.Target)
		}
	case "/":
		m.filtering = true
		return m, m.filter.Focus()
	case plumtypes.KEY_ESC:
		// Esc steps back like in the finder: clears an active filter, then the previous page, then quits.
		if m.filter.Value() != "" {
			m.filter.SetValue("")
			m.cursor, m.offset = 0, 0
			return m, nil
		}
		if len(m.history) > 0 {
			return m, m.back()
		}
		return m.quit(ActionQuit, Target{})
	case plumtypes.SHORT_CTRL_Q, plumtypes.SHORT_CTRL_C:
		return m.quit(ActionQuit, Target{})
	}
	return m, nil
}

func (m *model) View() tea.View {
	if m.action != ActionQuit || m.err != nil {
		return tea.NewView("") // the note view or a form takes the screen
	}

	rows := m.visible()
	fit := m.rows()
	m.offset = min(max(m.offset, m.cursor-fit+1), m.cursor)
	end := min(len(rows), m.offset+fit)

	empty := m.page.Empty
	if m.filter.Value() != "" {
		empty = "No match"
	}
	body := strings.TrimRight(Render(m.page, rows[m.offset:end], m.cursor-m.offset, empty), "\n")
	if m.filtering || m.filter.Value() != "" {
		body += "\n" + m.filter.View()
	}

	// Pad so the footer sits on the last line but one: the last stays for the shell prompt.
	padding := max(0, m.height-1-footerHeight-lipgloss.Height(body))
	return tea.NewView(body + strings.Repeat("\n", padding) + "\n\n" + m.footer())
}

// Helpers

// Render draws a page like show folder: bold title, faint path and info, then the rows, the selected one marked
// (cursor < 0: none, for a pipe).
func Render(p Page, rows []Row, cursor int, empty string) string {
	var output strings.Builder
	output.WriteString(lipgloss.NewStyle().Bold(true).Render(p.Title) + "\n")
	output.WriteString(lipgloss.NewStyle().Faint(true).Render(p.Path) + "\n")
	output.WriteString(lipgloss.NewStyle().Faint(true).Render(p.Info) + "\n\n")

	if len(rows) == 0 {
		output.WriteString(lipgloss.NewStyle().Faint(true).Italic(true).Render(empty))
		return output.String()
	}

	var lines strings.Builder
	table := tabwriter.NewWriter(&lines, 0, 0, 2, ' ', 0)
	for _, r := range rows {
		fmt.Fprintln(table, strings.Join(r.Cells, "\t"))
	}
	table.Flush() // writes to a strings.Builder: cannot fail

	// Styled after the tabwriter: escape codes would break its column widths.
	for i, line := range strings.Split(strings.TrimRight(lines.String(), "\n"), "\n") {
		switch {
		case cursor < 0:
			output.WriteString(line + "\n")
		case i == cursor:
			output.WriteString(lipgloss.NewStyle().Bold(true).Render("> "+line) + "\n")
		default:
			output.WriteString("  " + line + "\n")
		}
	}
	return output.String()
}

// visible is the page's rows matching the filter, case-insensitively, on any cell.
func (m *model) visible() []Row {
	query := strings.ToLower(strings.TrimSpace(m.filter.Value()))
	if query == "" {
		return m.page.Rows
	}

	var matches []Row
	for _, r := range m.page.Rows {
		if strings.Contains(strings.ToLower(strings.Join(r.Cells, " ")), query) {
			matches = append(matches, r)
		}
	}
	return matches
}

func (m *model) selected() (Row, bool) {
	rows := m.visible()
	if m.cursor < 0 || m.cursor >= len(rows) {
		return Row{}, false
	}
	return rows[m.cursor], true
}

// rows is how many rows fit between the header and the footer (and the filter line).
func (m *model) rows() int {
	if m.height == 0 {
		return len(m.page.Rows) + 1 // no size yet: show everything
	}
	filterHeight := 0
	if m.filtering || m.filter.Value() != "" {
		filterHeight = 1
	}
	return max(1, m.height-1-headerHeight-footerHeight-filterHeight)
}

func (m *model) open(target Target, back bool) tea.Cmd {
	return func() tea.Msg {
		page, err := m.load(target)
		return loadedMsg{target: target, page: page, back: back, err: err}
	}
}

// back reopens the previous page; nothing at the first one.
func (m *model) back() tea.Cmd {
	if len(m.history) == 0 {
		return nil
	}
	return m.open(m.history[len(m.history)-1], true)
}

func (m *model) quit(kind Action, target Target) (tea.Model, tea.Cmd) {
	m.action, m.target = kind, target
	return m, tea.Quit
}

// footer only offers open and edit when they apply: on a note row.
func (m *model) footer() string {
	enter := "Center"
	if m.options.EnterOpens {
		enter = "Open"
	}
	hints := []screen.Hint{{Key: plumtypes.KEY_ENTER, Label: enter}, {Key: plumtypes.KEY_ESC, Label: "Back"}}
	if selected, ok := m.selected(); ok && selected.Target.Kind == KindNote {
		if !m.options.EnterOpens { // ↳ already says Open
			hints = append(hints, screen.Hint{Key: plumtypes.SHORT_CTRL_O, Label: "Open"})
		}
		hints = append(hints, screen.Hint{Key: plumtypes.SHORT_CTRL_E, Label: "Edit"})
	}
	hints = append(hints, screen.Hint{Key: "/", Label: "Filter"}, screen.Hint{Key: plumtypes.SHORT_CTRL_Q, Label: "Quit"})
	return screen.Footer(hints...)
}
