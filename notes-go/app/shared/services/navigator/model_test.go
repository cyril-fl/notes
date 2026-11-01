package navigator

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// A graph-like set of pages: tag go is on notes A and B; A's neighbor is B.
var (
	tagsTarget = Target{Kind: "tags"}
	goTarget   = Target{Kind: "tag", Key: "go"}
	aTarget    = Target{Kind: KindNote, Key: "a"}
	bTarget    = Target{Kind: KindNote, Key: "b"}
	pages      = map[Target]Page{
		tagsTarget: {Title: "Graph", Rows: []Row{{Cells: []string{"go", "2 notes"}, Target: goTarget}}},
		goTarget:   {Title: "go", Rows: []Row{{Cells: []string{"A"}, Target: aTarget}, {Cells: []string{"B"}, Target: bTarget}}},
		aTarget:    {Title: "A", Rows: []Row{{Cells: []string{"B", "/B"}, Target: bTarget}}},
		bTarget:    {Title: "B", Empty: "alone"},
	}
)

func fakeLoad(target Target) (Page, error) {
	return pages[target], nil
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
	enter = tea.KeyPressMsg{Code: tea.KeyEnter}
	esc   = tea.KeyPressMsg{Code: tea.KeyEscape}
	ctrl  = func(r rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: r, Mod: tea.ModCtrl} }
	text  = func(r rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: r, Text: string(r)} }
)

func TestWalk(t *testing.T) {
	m := newModel(fakeLoad, Options{}, tagsTarget, pages[tagsTarget], nil, 0)

	press(m, enter) // tag go
	press(m, enter) // note A: centers, the graph way
	press(m, enter) // its neighbor B
	if m.current != bTarget || len(m.history) != 3 {
		t.Fatalf("current = %+v, history = %d; want B after 3 steps", m.current, len(m.history))
	}

	press(m, down) // empty page: the cursor stays at 0
	if m.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", m.cursor)
	}
	m.View()

	press(m, esc)
	press(m, esc)
	if m.current != goTarget || len(m.history) != 1 {
		t.Fatalf("after 2 esc: current = %+v, history = %d; want go, 1", m.current, len(m.history))
	}

	press(m, esc) // back to the first page
	if cmd := press(m, esc); cmd == nil || m.action != ActionQuit {
		t.Error("esc on the first page does not quit")
	}
}

func TestEnterOpens(t *testing.T) {
	// Search mode: ↳ on a note opens it instead of centering.
	m := newModel(fakeLoad, Options{EnterOpens: true}, goTarget, pages[goTarget], nil, 1)
	if cmd := press(m, enter); cmd == nil || m.action != ActionOpen || m.target != bTarget {
		t.Errorf("action = %v, target = %+v; want open B", m.action, m.target)
	}
}

func TestNoteActions(t *testing.T) {
	// Open and edit only apply to a note row.
	m := newModel(fakeLoad, Options{}, tagsTarget, pages[tagsTarget], nil, 0)
	if press(m, ctrl('o')); m.action != ActionQuit {
		t.Errorf("ctrl+o on a tag row: action = %v, want none", m.action)
	}

	for key, want := range map[rune]Action{'o': ActionOpen, 'e': ActionEdit} {
		m := newModel(fakeLoad, Options{}, goTarget, pages[goTarget], nil, 1)
		if cmd := press(m, ctrl(key)); cmd == nil || m.action != want || m.target != bTarget {
			t.Errorf("ctrl+%c: action = %v, target = %+v; want %v on B", key, m.action, m.target, want)
		}
	}
}

func TestFilter(t *testing.T) {
	m := newModel(fakeLoad, Options{}, goTarget, pages[goTarget], nil, 0)

	press(m, text('/'))
	press(m, text('b'))
	if rows := m.visible(); len(rows) != 1 || rows[0].Target != bTarget {
		t.Fatalf("visible = %+v, want [B]", rows)
	}

	press(m, enter)
	press(m, esc) // clears the filter, does not quit
	if m.filter.Value() != "" || m.action != ActionQuit || len(m.visible()) != 2 {
		t.Errorf("after esc: filter = %q, visible = %d", m.filter.Value(), len(m.visible()))
	}
}
