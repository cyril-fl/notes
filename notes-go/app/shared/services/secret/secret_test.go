package secret

import (
	"errors"
	"strings"
	"testing"
)

func TestSealOpen(t *testing.T) {
	account, err := newAEAD("dGVzdCBrZXkgb2YgYW4gYWNjb3VudCwgMzIgYnl0ZXM=")
	if err != nil {
		t.Fatal(err)
	}
	keys.Store("a", accountKey{gcm: account, blind: []byte("a")})
	keys.Store("c", accountKey{gcm: account, blind: []byte("c")})

	if _, err := Seal("b", "x"); !errors.Is(err, ErrNotLoaded) {
		t.Errorf("Seal(unloaded) err = %v, want ErrNotLoaded", err)
	}
	sealed, err := Seal("a", "# Idée #go")
	again, _ := Seal("a", "# Idée #go")
	if err != nil || !strings.HasPrefix(string(sealed), Prefix+"a:") || strings.Contains(string(sealed), "Idée") || sealed == again {
		t.Errorf("Seal() = %q, %q, %v; want two different values without the text", sealed, again, err)
	}
	if text, err := Open(string(sealed)); err != nil || text != "# Idée #go" {
		t.Errorf("Open() = %q, %v; want the text back", text, err)
	}
	if _, err := Open("written in clear"); err != ErrCorrupt {
		t.Errorf("Open(clear) err = %v, want ErrCorrupt", err)
	}
	if _, err := Open(string(sealed[:len(sealed)-4]) + "AAAA"); err != ErrCorrupt {
		t.Errorf("Open(tampered) err = %v, want ErrCorrupt", err)
	}
	if _, err := Text("in clear").Value(); err == nil {
		t.Error("Text(clear).Value() err = nil, want a refusal")
	}

	blind, err := Blind("a", "Idée")
	again2, _ := Blind("a", "IDÉE")
	other, _ := Blind("c", "Idée")
	if err != nil || blind != again2 || blind == other || strings.Contains(blind, "Id") {
		t.Errorf("Blind() = %q, %q, %q, %v; want the same value whatever the case, another for another account", blind, again2, other, err)
	}

	image := []byte("\x89PNG bytes")
	sealedImage, err := SealBytes("a", image)
	if err != nil || strings.Contains(string(sealedImage), "PNG") {
		t.Errorf("SealBytes() = %q, %v; want the bytes hidden", sealedImage, err)
	}
	if data, err := OpenBytes("a", sealedImage); err != nil || string(data) != string(image) {
		t.Errorf("OpenBytes() = %q, %v; want the bytes back", data, err)
	}
	if _, err := OpenBytes("a", image); err != ErrCorrupt {
		t.Errorf("OpenBytes(clear) err = %v, want ErrCorrupt", err)
	}
}

func TestRewrap(t *testing.T) {
	gcm, err := master()
	if err != nil {
		t.Fatalf("master() err = %v (NOTES_SECRET set?)", err)
	}
	const oldSecret = "b2xkIHNlY3JldCBvZiB0aGUgc2VydmVyLCAzMiBieXQ="
	old, _ := newAEAD(oldSecret)
	key := []byte("0123456789abcdef0123456789abcdef")

	rewrapped, err := Rewrap(seal(old, key), oldSecret)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := open(gcm, rewrapped); err != nil || string(got) != string(key) {
		t.Errorf("open(Rewrap()) = %q, %v; want the key, sealed by NOTES_SECRET", got, err)
	}
	if again, err := Rewrap(rewrapped, oldSecret); err != nil || again != rewrapped {
		t.Errorf("Rewrap(done) = %q, %v; want it unchanged", again, err)
	}
}
