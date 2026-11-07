package style

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestTableAlignsColumnsOnceUncolored(t *testing.T) {
	rows := [][]string{{"note", "Idées", "a"}, {"folder", "Go", "b"}}

	if got, want := ansi.Strip(Table(rows, 1, 1)), "  note    Idées  a\n› folder  Go     b\n"; got != want {
		t.Fatalf("Table() =\n%q\nwant\n%q", got, want)
	}
	if got, want := ansi.Strip(Table(rows, -1, 1)), "note    Idées  a\nfolder  Go     b\n"; got != want {
		t.Fatalf("Table(-1) =\n%q\nwant\n%q", got, want)
	}
}
