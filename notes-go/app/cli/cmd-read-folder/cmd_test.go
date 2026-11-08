package cmdreadfolder

import (
	"testing"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
)

func TestInfo(t *testing.T) {
	created := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)
	children := []domain.Item{{Kind: domain.FOLDER}, {Kind: domain.NOTE}, {Kind: domain.NOTE}}

	tests := map[string]struct {
		folder readitem.Result
		want   string
	}{
		"updated":      {readitem.Result{Item: domain.Item{CreatedAt: created, UpdatedAt: updated}, Children: children}, "1 folder · 2 notes · updated 2 Nov 2026"},
		"never edited": {readitem.Result{Item: domain.Item{CreatedAt: created}}, "0 folders · 0 notes · updated 1 Nov 2026"},
		"root":         {readitem.Result{Children: children[:1]}, "1 folder · 0 notes"},
	}

	for name, test := range tests {
		if got := info(test.folder); got != test.want {
			t.Errorf("%s: info() = %q, want %q", name, got, test.want)
		}
	}
}
