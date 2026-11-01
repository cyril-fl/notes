package screen

import (
	"testing"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

func TestViewKeys(t *testing.T) {
	tests := map[string]struct {
		key  tea.KeyPressMsg
		edit bool
	}{
		"ctrl+e edits": {tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl}, true},
		"q quits":      {tea.KeyPressMsg{Code: 'q', Text: "q"}, false},
		"esc quits":    {tea.KeyPressMsg{Code: tea.KeyEscape}, false},
	}

	for name, test := range tests {
		v := &view{content: viewport.New()}
		_, cmd := v.Update(test.key)
		if cmd == nil || !v.quitting || v.edit != test.edit {
			t.Errorf("%s: quitting = %v, edit = %v, want quitting with edit = %v", name, v.quitting, v.edit, test.edit)
		}
	}
}

func TestViewSize(t *testing.T) {
	v := &view{content: viewport.New()}
	v.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if got := v.content.Height(); got != 24-1-footerHeight {
		t.Errorf("viewport height = %d, want %d", got, 24-1-footerHeight)
	}
}
