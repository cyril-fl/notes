package secret

import (
	"strings"
	"testing"
)

func TestSealOpen(t *testing.T) {
	sealed, err := Seal("# Idée #go")
	if err != nil {
		t.Fatalf("Seal() err = %v (NOTES_SECRET set?)", err)
	}
	again, _ := Seal("# Idée #go")
	if !strings.HasPrefix(sealed, Prefix) || strings.Contains(sealed, "Idée") || sealed == again {
		t.Errorf("Seal() = %q, %q; want two different values without the text", sealed, again)
	}
	if text, err := Open(sealed); err != nil || text != "# Idée #go" {
		t.Errorf("Open() = %q, %v; want the text back", text, err)
	}
	if text, err := Open("written in clear"); err != nil || text != "written in clear" {
		t.Errorf("Open(clear) = %q, %v; want it as is", text, err)
	}
	if _, err := Open(sealed[:len(sealed)-4] + "AAAA"); err != ErrCorrupt {
		t.Errorf("Open(tampered) err = %v, want ErrCorrupt", err)
	}
}
