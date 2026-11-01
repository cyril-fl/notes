package millflagscreatenote

import (
	"errors"
	"os"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	plumtypes "github.com/cyril-fl/plum/types"
	"golang.org/x/term"
	"main.go/app/shared/domain"
)

var errRequired = errors.New("required")

// askMissing opens one form with only the fields the flags left empty; no field, no form.
func (m *mill) askMissing() {
	if m.HasErr() {
		return
	}

	var fields []huh.Field
	var content *huh.Text

	if m.cache.input.Title == "" && !m.cache.noTitle {
		fields = append(fields, huh.NewInput().Title("Title").Value(&m.cache.input.Title).Validate(notBlank))
	}
	if m.cache.input.Content == "" {
		content = huh.NewText().Title("Content").Value(&m.cache.input.Content).Validate(notBlank)
		fields = append(fields, content)
	}
	// No existing tag: nothing to pick, the list is skipped.
	if m.cache.tagFromList && len(m.cache.tags) > 0 {
		fields = append(fields, huh.NewMultiSelect[string]().Title("Tags").Options(tagOptions(m.cache.tags)...).Value(&m.cache.input.Tags).Filterable(true))
	}

	if len(fields) == 0 {
		return
	}

	if content != nil {
		fillTerminal(fields, content)
	}

	m.SetErr(newForm(fields).Run())
}

// Helpers

// fillTerminal grows the content field so the form takes the whole terminal height.
// A group sizes itself when built, hence the probe form to measure the other fields.
func fillTerminal(fields []huh.Field, content *huh.Text) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return
	}

	content.Lines(1)
	probe := newForm(fields).WithWidth(width)
	probe.Init()
	if lines := height - lipgloss.Height(probe.View()); lines > 1 {
		content.Lines(lines)
	}
}

func newForm(fields []huh.Field) *huh.Form {
	return huh.NewForm(huh.NewGroup(fields...)).WithKeyMap(keyMap()).WithTheme(huh.ThemeFunc(theme))
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
	t.Help.ShortKey = t.Help.ShortKey.Bold(true)
	t.Help.ShortSeparator = t.Help.ShortSeparator.Transform(func(string) string { return " · " })
	return t
}

func tagOptions(tags []domain.Tag) []huh.Option[string] {
	values := make([]string, len(tags))
	for i, tag := range tags {
		values[i] = tag.Value
	}

	return huh.NewOptions(values...)
}

func notBlank(value string) error {
	if strings.TrimSpace(value) == "" {
		return errRequired
	}

	return nil
}
