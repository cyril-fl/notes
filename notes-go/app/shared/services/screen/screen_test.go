package screen

import (
	"io"
	"os"
	"strings"
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

func TestPrintJSONEmptyList(t *testing.T) {
	stdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = writer
	t.Cleanup(func() { os.Stdout = stdout })

	PrintJSON[string](nil)
	writer.Close()

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(got)) != "[]" {
		t.Errorf("output = %q, want []", got)
	}
}
