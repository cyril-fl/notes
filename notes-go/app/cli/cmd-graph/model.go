package cmdgraph

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

// nodeKind is what the graph is centered on: the tag list, one tag, or one note.
type nodeKind int

const (
	nodeTags nodeKind = iota
	nodeTag
	nodeNote
)

// node is a place in the graph: key is the tag value or the note id.
type node struct {
	kind nodeKind
	key  string
}

// row is one line of a page; target is where ↳ goes.
type row struct {
	cells  []string
	target node
}

// page is what a node shows, rendered like show folder.
type page struct {
	title, path, info, empty string
	rows                     []row
}

// loader builds the page of a node.
type loader func(node) (page, error)

type loadedMsg struct {
	node node
	page page
	back bool
	err  error
}

// actionKind is what the graph hands over to cmd.go when it quits: another view or a form.
type actionKind int

const (
	actionQuit actionKind = iota
	actionOpen
	actionEdit
)

type model struct {
	load      loader
	node      node
	page      page
	history   []node
	cursor    int
	offset    int
	height    int
	filter    textinput.Model
	filtering bool
	action    actionKind
	target    node
	err       error
}

func newModel(load loader, current node, page page, history []node, cursor int) *model {
	m := &model{load: load, node: current, page: page, history: history, filter: textinput.New()}
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
			m.history = append(m.history, m.node)
		}
		m.node, m.page, m.cursor, m.offset = msg.node, msg.page, 0, 0
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
	isNote := hasSelected && selected.target.kind == nodeNote

	switch plumtypes.Key(msg.String()) {
	case plumtypes.KEY_UP, "k":
		m.cursor = max(0, m.cursor-1)
	case plumtypes.KEY_DOWN, "j":
		m.cursor = max(0, min(len(m.visible())-1, m.cursor+1)) // max: an empty list keeps the cursor at 0
	case plumtypes.KEY_ENTER, plumtypes.KEY_RIGHT:
		if hasSelected {
			return m, m.open(selected.target, false)
		}
	case plumtypes.KEY_LEFT, plumtypes.KEY_BACKSPACE:
		return m, m.back()
	case plumtypes.SHORT_CTRL_O:
		if isNote {
			return m.quit(actionOpen, selected.target)
		}
	case plumtypes.SHORT_CTRL_E:
		if isNote {
			return m.quit(actionEdit, selected.target)
		}
	case "/":
		m.filtering = true
		return m, m.filter.Focus()
	case plumtypes.KEY_ESC:
		// Esc steps back like in the finder: clears an active filter, then the previous center, then quits.
		if m.filter.Value() != "" {
			m.filter.SetValue("")
			m.cursor, m.offset = 0, 0
			return m, nil
		}
		if len(m.history) > 0 {
			return m, m.back()
		}
		return m.quit(actionQuit, node{})
	case plumtypes.SHORT_CTRL_Q, plumtypes.SHORT_CTRL_C:
		return m.quit(actionQuit, node{})
	}
	return m, nil
}

func (m *model) View() tea.View {
	if m.action != actionQuit || m.err != nil {
		return tea.NewView("") // the note view or a form takes the screen
	}

	rows := m.visible()
	fit := m.rows()
	m.offset = min(max(m.offset, m.cursor-fit+1), m.cursor)
	end := min(len(rows), m.offset+fit)

	empty := m.page.empty
	if m.filter.Value() != "" {
		empty = "No match"
	}
	body := strings.TrimRight(render(m.page, rows[m.offset:end], m.cursor-m.offset, empty), "\n")
	if m.filtering || m.filter.Value() != "" {
		body += "\n" + m.filter.View()
	}

	// Pad so the footer sits on the last line but one: the last stays for the shell prompt.
	padding := max(0, m.height-1-footerHeight-lipgloss.Height(body))
	return tea.NewView(body + strings.Repeat("\n", padding) + "\n\n" + m.footer())
}

// Helpers

// render draws a page like show folder: bold title, faint path and info, then the rows, the selected one marked.
func render(p page, rows []row, cursor int, empty string) string {
	var output strings.Builder
	output.WriteString(lipgloss.NewStyle().Bold(true).Render(p.title) + "\n")
	output.WriteString(lipgloss.NewStyle().Faint(true).Render(p.path) + "\n")
	output.WriteString(lipgloss.NewStyle().Faint(true).Render(p.info) + "\n\n")

	if len(rows) == 0 {
		output.WriteString(lipgloss.NewStyle().Faint(true).Italic(true).Render(empty))
		return output.String()
	}

	var lines strings.Builder
	table := tabwriter.NewWriter(&lines, 0, 0, 2, ' ', 0)
	for _, r := range rows {
		fmt.Fprintln(table, strings.Join(r.cells, "\t"))
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
func (m *model) visible() []row {
	query := strings.ToLower(strings.TrimSpace(m.filter.Value()))
	if query == "" {
		return m.page.rows
	}

	var matches []row
	for _, r := range m.page.rows {
		if strings.Contains(strings.ToLower(strings.Join(r.cells, " ")), query) {
			matches = append(matches, r)
		}
	}
	return matches
}

func (m *model) selected() (row, bool) {
	rows := m.visible()
	if m.cursor < 0 || m.cursor >= len(rows) {
		return row{}, false
	}
	return rows[m.cursor], true
}

// rows is how many rows fit between the header and the footer (and the filter line).
func (m *model) rows() int {
	if m.height == 0 {
		return len(m.page.rows) + 1 // no size yet: show everything
	}
	filterHeight := 0
	if m.filtering || m.filter.Value() != "" {
		filterHeight = 1
	}
	return max(1, m.height-1-headerHeight-footerHeight-filterHeight)
}

func (m *model) open(target node, back bool) tea.Cmd {
	return func() tea.Msg {
		page, err := m.load(target)
		return loadedMsg{node: target, page: page, back: back, err: err}
	}
}

// back reopens the previous center; nothing at the first one.
func (m *model) back() tea.Cmd {
	if len(m.history) == 0 {
		return nil
	}
	return m.open(m.history[len(m.history)-1], true)
}

func (m *model) quit(kind actionKind, target node) (tea.Model, tea.Cmd) {
	m.action, m.target = kind, target
	return m, tea.Quit
}

// footer only offers open and edit when they apply: on a note row.
func (m *model) footer() string {
	hints := []screen.Hint{{Key: plumtypes.KEY_ENTER, Label: "Center"}, {Key: plumtypes.KEY_ESC, Label: "Back"}}
	if selected, ok := m.selected(); ok && selected.target.kind == nodeNote {
		hints = append(hints, screen.Hint{Key: plumtypes.SHORT_CTRL_O, Label: "Open"}, screen.Hint{Key: plumtypes.SHORT_CTRL_E, Label: "Edit"})
	}
	hints = append(hints, screen.Hint{Key: "/", Label: "Filter"}, screen.Hint{Key: plumtypes.SHORT_CTRL_Q, Label: "Quit"})
	return screen.Footer(hints...)
}
