package cmdaccount

import "testing"

func TestSameName(t *testing.T) {
	check := sameName("cyril")
	if err := check("cyril"); err != nil {
		t.Errorf("the name: %v", err)
	}
	for _, typed := range []string{"", "Cyril", "cyril ", "cyri"} {
		if check(typed) == nil {
			t.Errorf("%q confirmed the purge", typed)
		}
	}
}
