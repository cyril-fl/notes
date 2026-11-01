package screen

import (
	"testing"

	"charm.land/lipgloss/v2"
)

func TestPad(t *testing.T) {
	if got := lipgloss.Height(pad("a\nb", 10)); got != 9 {
		t.Errorf("height = %d, want 9 (one line left for the prompt)", got)
	}
	if got := pad("a\nb\nc", 2); got != "a\nb\nc" {
		t.Errorf("pad() = %q, want the output unchanged", got)
	}
}
