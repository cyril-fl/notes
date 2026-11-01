package cmdtrash

import (
	"testing"
	"time"
)

func TestPurge(t *testing.T) {
	now := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	tests := map[string]struct {
		deletedAt time.Time
		retention int
		want      string
	}{
		"today":      {now, 30, "purged in 30 days"},
		"last day":   {now.AddDate(0, 0, -29), 30, "purged in 1 day"},
		"overdue":    {now.AddDate(0, 0, -31), 30, "purged at next start"},
		"never":      {now.AddDate(0, 0, -400), 0, "kept"},
		"some hours": {now.Add(-5 * time.Hour), 2, "purged in 2 days"},
	}

	for name, test := range tests {
		if got := purge(test.deletedAt, test.retention, now); got != test.want {
			t.Errorf("%s: purge() = %q, want %q", name, got, test.want)
		}
	}
}
