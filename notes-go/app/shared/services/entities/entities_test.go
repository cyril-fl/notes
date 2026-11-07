package entities

import "testing"

func TestDecode(t *testing.T) {
	cases := map[string]string{
		"a &gt;&gt; b &amp; c &lt; d": "a >> b & c < d",
		"&amp;lt; et &amp;amp;":       "&lt; et &amp;",
		"`&lt;` &lt;":                 "`&lt;` <",
		"```\n&amp;\n```\n&amp;":      "```\n&amp;\n```\n&",
		"~~~go\na &lt; b\n~~~":        "~~~go\na &lt; b\n~~~",
		"rien à faire":                "rien à faire",
	}
	for in, want := range cases {
		if got := Decode(in); got != want {
			t.Errorf("Decode(%q) = %q, want %q", in, got, want)
		}
	}
}
