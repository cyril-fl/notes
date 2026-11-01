package cmdgraph

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// The graph: tag go is on notes A and B; A's neighbor is B.
var (
	tagsNode = node{kind: nodeTags}
	goNode   = node{kind: nodeTag, key: "go"}
	aNode    = node{kind: nodeNote, key: "a"}
	bNode    = node{kind: nodeNote, key: "b"}
	pages    = map[node]page{
		tagsNode: {title: "Graph", rows: []row{{cells: []string{"go", "2 notes"}, target: goNode}}},
		goNode:   {title: "go", rows: []row{{cells: []string{"A"}, target: aNode}, {cells: []string{"B"}, target: bNode}}},
		aNode:    {title: "A", rows: []row{{cells: []string{"B", "/B"}, target: bNode}}},
		bNode:    {title: "B", empty: "alone"},
	}
)

func fakeLoad(n node) (page, error) {
	return pages[n], nil
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
	m := newModel(fakeLoad, tagsNode, pages[tagsNode], nil, 0)

	press(m, enter) // tag go
	press(m, enter) // note A
	press(m, enter) // its neighbor B
	if m.node != bNode || len(m.history) != 3 {
		t.Fatalf("node = %+v, history = %d; want B after 3 steps", m.node, len(m.history))
	}

	press(m, down) // empty page: the cursor stays at 0
	if m.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", m.cursor)
	}
	m.View()

	press(m, esc)
	press(m, esc)
	if m.node != goNode || len(m.history) != 1 {
		t.Fatalf("after 2 esc: node = %+v, history = %d; want go, 1", m.node, len(m.history))
	}

	press(m, esc) // back to the tag list
	if cmd := press(m, esc); cmd == nil || m.action != actionQuit {
		t.Error("esc on the first center does not quit")
	}
}

func TestNoteActions(t *testing.T) {
	// Open and edit only apply to a note row.
	m := newModel(fakeLoad, tagsNode, pages[tagsNode], nil, 0)
	if press(m, ctrl('o')); m.action != actionQuit {
		t.Errorf("ctrl+o on a tag row: action = %v, want none", m.action)
	}

	for key, want := range map[rune]actionKind{'o': actionOpen, 'e': actionEdit} {
		m := newModel(fakeLoad, goNode, pages[goNode], nil, 1)
		if cmd := press(m, ctrl(key)); cmd == nil || m.action != want || m.target != bNode {
			t.Errorf("ctrl+%c: action = %v, target = %+v; want %v on B", key, m.action, m.target, want)
		}
	}
}

func TestFilter(t *testing.T) {
	m := newModel(fakeLoad, goNode, pages[goNode], nil, 0)

	press(m, text('/'))
	press(m, text('b'))
	if rows := m.visible(); len(rows) != 1 || rows[0].target != bNode {
		t.Fatalf("visible = %+v, want [B]", rows)
	}

	press(m, enter)
	press(m, esc) // clears the filter, does not quit
	if m.filter.Value() != "" || m.action != actionQuit || len(m.visible()) != 2 {
		t.Errorf("after esc: filter = %q, visible = %d", m.filter.Value(), len(m.visible()))
	}
}
