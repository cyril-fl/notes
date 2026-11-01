package millflagseditfolder

import (
	"errors"
	"strings"

	"charm.land/huh/v2"
	"main.go/app/shared/services/huhform"
)

var errRequired = errors.New("required")

// askMissing asks for the title, pre-filled, when --title is not given.
func (m *mill) askMissing() {
	if m.HasErr() || m.flags.Changed("title") {
		return
	}

	m.SetErr(huhform.New(huh.NewInput().Title("Title").Value(&m.cache.input.Title).Validate(notBlank)).Run())
}

// Helpers

func notBlank(value string) error {
	if strings.TrimSpace(value) == "" {
		return errRequired
	}

	return nil
}
