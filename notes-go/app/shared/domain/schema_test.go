package domain

import (
	"slices"
	"testing"
)

func TestCompareFold(t *testing.T) {
	got := []string{"zoo", "Été", "abricot", "Abricot", "été", "ete"}
	slices.SortStableFunc(got, CompareFold)
	want := []string{"abricot", "Abricot", "ete", "Été", "été", "zoo"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
