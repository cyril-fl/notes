package millflagscreatefolder

import (
	"errors"
	"strings"

	"charm.land/huh/v2"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/huhform"
)

var errRequired = errors.New("required")

// askMissing asks for the title when --title is not given; no missing field, no form.
func (m *mill) askMissing() {
	if m.HasErr() || m.cache.input.Title != "" {
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
