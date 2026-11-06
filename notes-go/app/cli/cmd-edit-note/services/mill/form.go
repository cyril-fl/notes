package millflagseditnote

import (
	"errors"
	"strings"

	"charm.land/huh/v2"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/hashtag"
	"main.go/app/shared/services/huhform"
)

var errRequired = errors.New("required")

// askMissing opens one form pre-filled with the current values: the content when no flag
// gave a change, the tag list with --tag-from-list; no field, no form.
func (m *mill) askMissing() {
	if m.HasErr() {
		return
	}

	var fields []huh.Field

	if !m.hasChanges() {
		fields = append(fields, huh.NewText().Title("Content").Value(&m.cache.input.Content).Validate(notBlank))
	}
	if m.cache.tagFromList {
		options := tagOptions(m.cache.tags, m.cache.input.Tags)
		fields = append(fields, huh.NewMultiSelect[string]().Title("Tags").Options(huh.NewOptions(options...)...).Value(&m.cache.input.Tags).Filterable(true))
	}

	if len(fields) == 0 {
		return
	}

	m.SetErr(huhform.New(fields...).Run())
	if !m.cache.tagFromList {
		m.cache.input.Tags = hashtag.Parse(m.cache.input.Content) // The form may have changed the text.
	}
}

// Helpers

// tagOptions lists the existing tags plus the note's tags not saved yet (--tag), and rewrites
// selected to the existing spelling: the multiselect only checks exact matches.
func tagOptions(tags []domain.Tag, selected []string) []string {
	options := make([]string, len(tags))
	for i, tag := range tags {
		options[i] = tag.Value
	}

	for i, value := range selected {
		found := false
		for _, option := range options {
			if strings.EqualFold(option, value) {
				selected[i], found = option, true
				break
			}
		}
		if !found {
			options = append(options, value)
		}
	}

	return options
}

func notBlank(value string) error {
	if strings.TrimSpace(value) == "" {
		return errRequired
	}

	return nil
}
