// Package huhform gives every CLI form the same keys and SCNe-like help line.
package huhform

import (
	"os"

	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	plumtypes "github.com/cyril-fl/plum/types"
	"golang.org/x/term"
	"main.go/app/shared/services/style"
)

// New builds a one-group form with the shared key map and theme, as tall as the terminal.
// The help line always shows so the form keeps its height when an error replaces it.
func New(fields ...huh.Field) *huh.Form {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return build(fields)
	}
	return fit(fields, width, height)
}

// fit grows the last text field to fill the height, or pads the form when there is none.
// A group sizes itself when built, hence the probe form to measure the other fields.
func fit(fields []huh.Field, width, height int) *huh.Form {
	text := lastText(fields)
	if text != nil {
		text.Lines(1)
	}

	probe := build(fields).WithWidth(width)
	probe.Init()
	extra := height - lipgloss.Height(probe.View())
	if extra <= 0 {
		return build(fields)
	}
	if text != nil {
		text.Lines(extra) // not 1+extra: measured, the 1-line probe already counts one extra row
		return build(fields)
	}
	// WithHeight sizes the group, the help line comes on top of it.
	return build(fields).WithHeight(height - 1)
}

func build(fields []huh.Field) *huh.Form {
	return huh.NewForm(huh.NewGroup(fields...)).WithKeyMap(keyMap()).WithTheme(huh.ThemeFunc(theme))
}

func lastText(fields []huh.Field) *huh.Text {
	for i := len(fields) - 1; i >= 0; i-- {
		if text, ok := fields[i].(*huh.Text); ok {
			return text
		}
	}
	return nil
}

// keyMap makes the content field behave like an editor: enter adds a line, ctrl+enter moves on.
// tab stays as a fallback for terminals that cannot tell ctrl+enter from enter.
func keyMap() *huh.KeyMap {
	keys := huh.NewDefaultKeyMap()
	keys.Input.Prev = shortcut("Back", plumtypes.KEY_SHIFT_TAB)
	keys.Input.Next = shortcut("Next", plumtypes.KEY_ENTER, plumtypes.KEY_TAB)
	keys.Input.Submit = shortcut("Submit", plumtypes.KEY_ENTER)
	keys.Text.Prev = shortcut("Back", plumtypes.KEY_SHIFT_TAB)
	keys.Text.NewLine = shortcut("New line", plumtypes.KEY_ENTER)
	keys.Text.Editor = shortcut("Editor", plumtypes.SHORT_CTRL_E)
	keys.Text.Next = shortcut("Next", plumtypes.CMD_CTRL_ENTER, plumtypes.KEY_TAB)
	keys.Text.Submit = shortcut("Submit", plumtypes.CMD_CTRL_ENTER, plumtypes.KEY_TAB)
	return keys
}

// shortcut binds the keys and shows the first one as a glyph in the help line ("^↳ Submit").
func shortcut(label string, keys ...plumtypes.Key) key.Binding {
	names := make([]string, len(keys))
	for i, k := range keys {
		names[i] = string(k)
	}
	return key.NewBinding(key.WithKeys(names...), key.WithHelp(keys[0].Symbol(), label))
}

// theme renders the help line like the SCNe footer: bold key, muted label, " · " between.
func theme(isDark bool) *huh.Styles {
	t := huh.ThemeCharm(isDark)
	t.Help.ShortKey = style.Key()
	t.Help.ShortDesc = style.Muted()
	t.Help.ShortSeparator = style.Subtle().Transform(func(string) string { return " · " })
	t.Focused.Title = t.Focused.Title.Foreground(style.ColorAccent())
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(style.ColorAccent())
	t.Focused.Base = t.Focused.Base.BorderForeground(style.ColorAccent())
	return t
}
