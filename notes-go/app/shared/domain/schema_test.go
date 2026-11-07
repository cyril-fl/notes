package domain

import (
	"slices"
	"testing"
)

func TestSortFold(t *testing.T) {
	got := []string{"zoo", "Été", "abricot", "Abricot", "été", "ete"}
	SortFold(got, func(s string) string { return s }, nil)
	want := []string{"abricot", "Abricot", "ete", "Été", "été", "zoo"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	// first wins over the key: the longest first, then by name.
	got = []string{"zoo", "Été", "abricot"}
	SortFold(got, func(s string) string { return s }, func(a, b string) int { return len(b) - len(a) })
	want = []string{"abricot", "Été", "zoo"}
	if !slices.Equal(got, want) {
		t.Errorf("with first: got %v, want %v", got, want)
	}
}
