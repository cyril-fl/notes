package cmdfinder

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
)

// tree: root holds folder A (with note N inside) and note R.
var (
	rootID    = ""
	folderA   = domain.Item{ID: "a", Kind: domain.FOLDER, Title: "A"}
	noteR     = domain.Item{ID: "r", Kind: domain.NOTE, Title: "Readme"}
	noteN     = domain.Item{ID: "n", Kind: domain.NOTE, Title: "Notes", ParentID: &folderA.ID}
	folderOfA = readitem.Result{Item: domain.Item{ID: "a", Kind: domain.FOLDER, Title: "A", ParentID: &rootID}, Path: "/A", Children: []domain.Item{noteN}}
	root      = readitem.Result{Item: domain.Item{Kind: domain.FOLDER, Title: "Notes"}, Path: "/", Children: []domain.Item{folderA, noteR}}
)

func fakeLoad(id string) (readitem.Result, error) {
	if id == "a" {
		return folderOfA, nil
	}
	return root, nil
}

// press sends a key and runs the command it returns, like the bubbletea loop would.
func press(m *model, key tea.KeyPressMsg) tea.Cmd {
	_, cmd := m.Update(key)
	if cmd != nil {
		if msg, ok := cmd().(loadedMsg); ok {
			m.Update(msg)
			return nil
		}
	}
	return cmd
}

var (
	down  = tea.KeyPressMsg{Code: tea.KeyDown}
	up    = tea.KeyPressMsg{Code: tea.KeyUp}
	enter = tea.KeyPressMsg{Code: tea.KeyEnter}
	left  = tea.KeyPressMsg{Code: tea.KeyLeft}
	esc   = tea.KeyPressMsg{Code: tea.KeyEscape}
	ctrl  = func(r rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: r, Mod: tea.ModCtrl} }
	text  = func(r rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: r, Text: string(r)} }
)

func TestNavigate(t *testing.T) {
	m := newModel(fakeLoad, root, 0, nil, "", false)

	press(m, up)
	press(m, down)
	press(m, down)
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1 (bounded)", m.cursor)
	}

	press(m, up)
	press(m, enter)
	if m.folder.Path != "/A" || m.cursor != 0 {
		t.Fatalf("after enter: path = %q, cursor = %d; want /A, 0", m.folder.Path, m.cursor)
	}

	press(m, left)
	if m.folder.Path != "/" || m.cursor != 0 {
		t.Errorf("after back: path = %q, cursor = %d; want /, cursor on A", m.folder.Path, m.cursor)
	}

	press(m, enter)
	press(m, esc) // Esc goes back too
	if m.folder.Path != "/" || m.action.kind != actionQuit {
		t.Errorf("after esc in /A: path = %q, action = %v; want /, still running", m.folder.Path, m.action.kind)
	}

	press(m, left) // no-op at the root
	if m.folder.Path != "/" || m.action.kind != actionQuit {
		t.Errorf("back at root: path = %q, action = %v", m.folder.Path, m.action.kind)
	}
}

func TestActions(t *testing.T) {
	tests := map[string]struct {
		cursor int
		key    tea.KeyPressMsg
		want   action
	}{
		"enter on note opens": {1, enter, action{actionOpen, noteR}},
		"ctrl+e edits":        {0, ctrl('e'), action{actionEdit, folderA}},
		"ctrl+n creates":      {0, ctrl('n'), action{actionCreate, domain.Item{}}},
		"ctrl+d deletes":      {1, ctrl('d'), action{actionDelete, noteR}},
		"ctrl+q quits":        {0, ctrl('q'), action{actionQuit, domain.Item{}}},
	}

	for name, test := range tests {
		m := newModel(fakeLoad, root, test.cursor, nil, "", false)
		if cmd := press(m, test.key); cmd == nil {
			t.Errorf("%s: no quit command", name)
		}
		if m.action.kind != test.want.kind || m.action.item.ID != test.want.item.ID {
			t.Errorf("%s: action = %+v, want %+v", name, m.action, test.want)
		}
	}
}

func TestFilter(t *testing.T) {
	m := newModel(fakeLoad, root, 0, nil, "", false)

	press(m, text('/'))
	for _, r := range "READ" {
		press(m, text(r))
	}
	if got := m.visible(); len(got) != 1 || got[0].ID != noteR.ID {
		t.Fatalf("visible = %+v, want [Readme]", got)
	}

	press(m, enter) // keeps the filter, leaves typing
	if m.filtering || m.filter.Value() == "" {
		t.Fatalf("filtering = %v, filter = %q; want kept, not typing", m.filtering, m.filter.Value())
	}

	// At the root, the first Esc clears the filter, the second quits.
	if press(m, esc); m.filter.Value() != "" || m.action.kind != actionQuit || len(m.visible()) != 2 {
		t.Fatalf("after esc: filter = %q, visible = %d", m.filter.Value(), len(m.visible()))
	}
	if cmd := press(m, esc); cmd == nil {
		t.Error("second esc does not quit")
	}
}

func TestScroll(t *testing.T) {
	folder := readitem.Result{Path: "/"}
	for i := 0; i < 20; i++ {
		folder.Children = append(folder.Children, domain.Item{ID: string(rune('a' + i)), Kind: domain.NOTE, Title: secret.Text(rune('a' + i))})
	}
	m := newModel(fakeLoad, folder, 0, nil, "", false)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 12}) // 12 - 1 - 3 - 2 = 6 rows

	for i := 0; i < 10; i++ {
		press(m, down)
	}
	m.View()
	if m.cursor < m.offset || m.cursor >= m.offset+m.rows() {
		t.Errorf("cursor %d outside the window [%d, %d)", m.cursor, m.offset, m.offset+m.rows())
	}
}

// Regression: down in an empty folder set the cursor to -1 and View panicked on children[-1:].
func TestEmptyFolder(t *testing.T) {
	m := newModel(fakeLoad, readitem.Result{Path: "/"}, 0, nil, "", false)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 12})

	press(m, down)
	press(m, up)
	if m.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", m.cursor)
	}
	m.View()

	if press(m, enter); m.action.kind != actionQuit {
		t.Errorf("enter on nothing: action = %v, want none", m.action.kind)
	}
}

func TestCutPaste(t *testing.T) {
	m := newModel(fakeLoad, root, 1, nil, "", false)

	if press(m, ctrl('v')); m.action.kind != actionQuit {
		t.Fatalf("paste with nothing cut: action = %v, want none", m.action.kind)
	}

	press(m, ctrl('x')) // cut Readme
	if m.clipboard == nil || m.clipboard.ID != noteR.ID {
		t.Fatalf("clipboard = %+v, want Readme", m.clipboard)
	}

	press(m, esc) // cancels the cut, does not quit
	if m.clipboard != nil || m.action.kind != actionQuit {
		t.Fatalf("after esc: clipboard = %+v, action = %v", m.clipboard, m.action.kind)
	}

	press(m, ctrl('x'))
	press(m, up)
	press(m, enter) // into A, the clipboard follows
	if cmd := press(m, ctrl('v')); cmd == nil || m.action.kind != actionMove || m.action.item.ID != noteR.ID || m.folder.Path != "/A" {
		t.Errorf("paste: action = %+v, folder = %q; want move Readme into /A", m.action, m.folder.Path)
	}
}
