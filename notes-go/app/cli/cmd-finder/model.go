package cmdfinder

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	plumtypes "github.com/cyril-fl/plum/types"
	cmdreadfolder "main.go/app/cli/cmd-read-folder"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/process/readitem"
	"main.go/app/shared/services/screen"
	"main.go/app/shared/services/style"
)

// headerHeight is the title, path, info and blank line above the children; footerHeight the blank and shortcut lines.
const (
	headerHeight = 4
	footerHeight = 2
)

// actionKind is what the finder hands over to cmd.go when it quits: a form or another view.
type actionKind int

const (
	actionQuit actionKind = iota
	actionOpen
	actionEdit
	actionCreate
	actionDelete
	actionMove
)

type action struct {
	kind actionKind
	item domain.Item
}

// loader reads a folder with its children; "" is the root.
type loader func(id string) (readitem.Result, error)

// loadedMsg carries a folder read in the background; selectID puts the cursor on that child.
type loadedMsg struct {
	folder   readitem.Result
	selectID string
	err      error
}

type model struct {
	load      loader
	folder    readitem.Result
	cursor    int
	offset    int
	height    int
	filter    textinput.Model
	filtering bool
	clipboard *domain.Item // cut with ^X, pasted with ^V; kept by cmd.go across screens
	status    string       // result of the last action, shown above the footer
	failed    bool         // status is an error
	action    action
	err       error
}

func newModel(load loader, folder readitem.Result, cursor int, clipboard *domain.Item, status string, failed bool) *model {
	m := &model{load: load, folder: folder, clipboard: clipboard, status: status, failed: failed, filter: textinput.New()}
	m.filter.Prompt = "/"
	m.filter.SetStyles(style.Filter(m.filter.Styles()))
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
		m.folder, m.cursor, m.offset = msg.folder, 0, 0
		m.filter.SetValue("")
		for i, child := range m.visible() {
			if child.ID == msg.selectID {
				m.cursor = i
			}
		}
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

	switch plumtypes.Key(msg.String()) {
	case plumtypes.KEY_UP, "k":
		m.cursor = max(0, m.cursor-1)
	case plumtypes.KEY_DOWN, "j":
		m.cursor = max(0, min(len(m.visible())-1, m.cursor+1)) // max: an empty list keeps the cursor at 0
	case plumtypes.KEY_ENTER, plumtypes.KEY_RIGHT:
		switch {
		case !hasSelected:
		case selected.Kind == domain.FOLDER:
			return m, m.open(selected.ID, "")
		default:
			return m.quit(actionOpen, selected)
		}
	case plumtypes.KEY_LEFT, plumtypes.KEY_BACKSPACE:
		return m, m.back()
	case plumtypes.SHORT_CTRL_E:
		if hasSelected {
			return m.quit(actionEdit, selected)
		}
	case plumtypes.SHORT_CTRL_N:
		return m.quit(actionCreate, domain.Item{})
	case plumtypes.SHORT_CTRL_D:
		if hasSelected {
			return m.quit(actionDelete, selected)
		}
	case plumtypes.SHORT_CTRL_X:
		if hasSelected {
			m.clipboard, m.status, m.failed = &selected, "", false
		}
	case plumtypes.SHORT_CTRL_V:
		if m.clipboard != nil {
			return m.quit(actionMove, *m.clipboard)
		}
	case "/":
		m.filtering = true
		return m, m.filter.Focus()
	case plumtypes.KEY_ESC:
		// Esc steps back like in the note view: clears an active filter, then the clipboard, then goes up, then quits at the root.
		if m.filter.Value() != "" {
			m.filter.SetValue("")
			m.cursor, m.offset = 0, 0
			return m, nil
		}
		if m.clipboard != nil {
			m.clipboard = nil
			return m, nil
		}
		if m.folder.Item.ID != "" {
			return m, m.back()
		}
		return m.quit(actionQuit, domain.Item{})
	case plumtypes.SHORT_CTRL_Q, plumtypes.SHORT_CTRL_C:
		return m.quit(actionQuit, domain.Item{})
	}
	return m, nil
}

func (m *model) View() tea.View {
	if m.action.kind != actionQuit || m.err != nil {
		return tea.NewView("") // a form or the note view takes the screen
	}

	children := m.visible()
	rows := m.rows()
	m.offset = min(max(m.offset, m.cursor-rows+1), m.cursor)
	end := min(len(children), m.offset+rows)

	empty := cmdreadfolder.EmptyFolder
	if m.filter.Value() != "" {
		empty = "No match"
	}
	body := strings.TrimRight(cmdreadfolder.Output(m.folder, children[m.offset:end], m.cursor-m.offset, empty), "\n")
	if m.filtering || m.filter.Value() != "" {
		body += "\n" + m.filter.View()
	}
	if line := m.statusLine(); line != "" {
		body += "\n\n" + line
	}

	// Pad so the footer sits on the last line but one: the last stays for the shell prompt.
	padding := max(0, m.height-1-footerHeight-lipgloss.Height(body))
	return tea.NewView(body + strings.Repeat("\n", padding) + "\n\n" + m.footer())
}

// Helpers

// visible is the folder's children matching the filter, case-insensitively.
func (m *model) visible() []domain.Item {
	query := strings.ToLower(strings.TrimSpace(m.filter.Value()))
	if query == "" {
		return m.folder.Children
	}

	var matches []domain.Item
	for _, child := range m.folder.Children {
		if strings.Contains(strings.ToLower(string(child.Title)), query) {
			matches = append(matches, child)
		}
	}
	return matches
}

func (m *model) selected() (domain.Item, bool) {
	children := m.visible()
	if m.cursor < 0 || m.cursor >= len(children) {
		return domain.Item{}, false
	}
	return children[m.cursor], true
}

// rows is how many children fit between the header and the footer (and the filter line).
func (m *model) rows() int {
	if m.height == 0 {
		return len(m.folder.Children) + 1 // no size yet: show everything
	}
	extra := 0
	if m.filtering || m.filter.Value() != "" {
		extra++
	}
	if m.statusLine() != "" {
		extra += 2
	}
	return max(1, m.height-1-headerHeight-footerHeight-extra)
}

func (m *model) open(id string, selectID string) tea.Cmd {
	return func() tea.Msg {
		folder, err := m.load(id)
		return loadedMsg{folder: folder, selectID: selectID, err: err}
	}
}

// back opens the parent folder with the cursor on the folder we leave; nothing at the root.
func (m *model) back() tea.Cmd {
	if m.folder.Item.ID == "" {
		return nil
	}
	parentID := ""
	if m.folder.Item.ParentID != nil {
		parentID = *m.folder.Item.ParentID
	}
	return m.open(parentID, m.folder.Item.ID)
}

func (m *model) quit(kind actionKind, item domain.Item) (tea.Model, tea.Cmd) {
	m.action = action{kind: kind, item: item}
	return m, tea.Quit
}

// statusLine shows a refused action first (the cut item stays, the footer still offers ^V),
// then the cut item waiting to be pasted, then the result of the last action.
func (m *model) statusLine() string {
	switch {
	case m.failed:
		return style.Error().Render(m.status)
	case m.clipboard != nil:
		return style.Muted().Render(fmt.Sprintf("✂ %s · ^V pastes it here, ⎋ cancels", untitled(string(m.clipboard.Title))))
	case m.status != "":
		return style.Muted().Render(m.status)
	}
	return ""
}

func (m *model) footer() string {
	hints := []screen.Hint{
		{Key: plumtypes.KEY_ENTER, Label: "Open"},
		{Key: plumtypes.KEY_ESC, Label: "Back"},
		{Key: plumtypes.SHORT_CTRL_E, Label: "Edit"},
		{Key: plumtypes.SHORT_CTRL_N, Label: "New"},
		{Key: plumtypes.SHORT_CTRL_D, Label: "Delete"},
		{Key: plumtypes.SHORT_CTRL_X, Label: "Cut"},
	}
	if m.clipboard != nil {
		hints = append(hints, screen.Hint{Key: plumtypes.SHORT_CTRL_V, Label: "Paste"})
	}
	return screen.Footer(append(hints, screen.Hint{Key: "/", Label: "Filter"}, screen.Hint{Key: plumtypes.SHORT_CTRL_Q, Label: "Quit"})...)
}

func untitled(title string) string {
	if title == "" {
		return "untitled"
	}
	return title
}
