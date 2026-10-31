package millflagscreatefolder

import (
	"errors"
	"strings"

	"charm.land/huh/v2"
)

var errRequired = errors.New("required")

// askMissing asks for the title when --title is not given; no missing field, no form.
func (m *mill) askMissing() {
	if m.HasErr() || m.cache.input.Title != "" {
		return
	}

	m.SetErr(huh.NewInput().Title("Title").Value(&m.cache.input.Title).Validate(notBlank).Run())
}

// Helpers

func notBlank(value string) error {
	if strings.TrimSpace(value) == "" {
		return errRequired
	}

	return nil
}
