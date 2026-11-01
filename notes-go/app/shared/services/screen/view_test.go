package screen

import (
	"testing"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

func TestViewKeys(t *testing.T) {
	tests := map[string]struct {
		key        tea.KeyPressMsg
		edit, quit bool
	}{
		"ctrl+e edits":   {tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl}, true, false},
		"esc goes back":  {tea.KeyPressMsg{Code: tea.KeyEscape}, false, false},
		"ctrl+q quits":   {tea.KeyPressMsg{Code: 'q', Mod: tea.ModCtrl}, false, true},
		"q is just text": {tea.KeyPressMsg{Code: 'q', Text: "q"}, false, false},
	}

	for name, test := range tests {
		v := &view{content: viewport.New()}
		v.Update(test.key)
		leaves := test.edit || test.quit || name == "esc goes back"
		if v.quitting != leaves || v.edit != test.edit || v.quit != test.quit {
			t.Errorf("%s: quitting = %v, edit = %v, quit = %v", name, v.quitting, v.edit, v.quit)
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
