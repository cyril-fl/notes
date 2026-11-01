package cmdeditfolder

import (
	"testing"

	"github.com/spf13/pflag"
)

// Interactive relies on it: a changed flag would skip the form.
func TestNoFlags(t *testing.T) {
	flags := noFlags()
	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Changed {
			t.Errorf("--%s is changed", flag.Name)
		}
	})
	if flags.Lookup("title") == nil {
		t.Error("--title is not declared")
	}
}
