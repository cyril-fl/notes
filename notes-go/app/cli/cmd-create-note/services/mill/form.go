package millflagscreatenote

import (
	"errors"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
	"main.go/app/shared/domain"
)

var errRequired = errors.New("required")

// askMissing opens one form with only the fields the flags left empty; no field, no form.
func (m *mill) askMissing() {
	if m.HasErr() {
		return
	}

	var fields []huh.Field

	if m.cache.input.Title == "" && !m.cache.noTitle {
		fields = append(fields, huh.NewInput().Title("Title").Value(&m.cache.input.Title).Validate(notBlank))
	}
	if m.cache.input.Content == "" {
		fields = append(fields, huh.NewText().Title("Content").Value(&m.cache.input.Content).Validate(notBlank))
	}
	// No existing tag: nothing to pick, the list is skipped.
	if m.cache.tagFromList && len(m.cache.tags) > 0 {
		fields = append(fields, huh.NewMultiSelect[string]().Title("Tags").Options(tagOptions(m.cache.tags)...).Value(&m.cache.input.Tags).Filterable(true))
	}

	if len(fields) == 0 {
		return
	}

	m.SetErr(huh.NewForm(huh.NewGroup(fields...)).WithKeyMap(keyMap()).Run())
}

// Helpers

// keyMap keeps alt+enter as the only new line key in the content field (huh also binds ctrl+j).
func keyMap() *huh.KeyMap {
	keys := huh.NewDefaultKeyMap()
	keys.Text.NewLine = key.NewBinding(key.WithKeys("alt+enter"), key.WithHelp("alt+enter", "new line"))
	return keys
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
