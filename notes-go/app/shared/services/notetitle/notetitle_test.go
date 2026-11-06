package notetitle

import (
	"strings"
	"testing"
)

func TestFrom(t *testing.T) {
	long := strings.Repeat("mot ", 20)
	cases := map[string]struct {
		content string
		want    string
	}{
		"heading":              {"# Courses\n\ndu pain", "Courses"},
		"first heading":        {"intro\n\n## Plan\n\n# Plus bas", "Plan"},
		"closing hashes":       {"## Plan ##", "Plan"},
		"heading marks":        {"# Les **bases** de [Go](http://go.dev)", "Les bases de Go"},
		"tags in heading":      {"# Courses #maison", "Courses"},
		"empty heading":        {"#\n\ntexte", "texte"},
		"first line":           {"du pain\net du lait", "du pain"},
		"skips tag lines":      {"#go #sql\n\nidée", "idée"},
		"skips blank lines":    {"\n\n   \ntexte", "texte"},
		"list item":            {"- [ ] acheter du pain", "acheter du pain"},
		"numbered":             {"1. un", "un"},
		"quote":                {"> cité", "cité"},
		"image":                {"![chat](a.png)", "chat"},
		"note link":            {"voir [[Courses]]", "voir Courses"},
		"resolved note link":   {"voir [Courses](note:11111111-1111-1111-1111-111111111111)", "voir Courses"},
		"rule":                 {"---\n\ntexte", "texte"},
		"code is not heading":  {"```sh\n# comment\n```\ntexte", "texte"},
		"only code":            {"```\ncode\n```", ""},
		"cut":                  {long, strings.TrimSpace(long[:MaxLength])},
		"cut runes":            {strings.Repeat("é", 60), strings.Repeat("é", MaxLength)},
		"only tags":            {"#go", ""},
		"empty":                {"", ""},
		"not a heading":        {"#hashtag\n#titre", ""},
		"heading over 50 kept": {"# " + long, strings.TrimSpace(long)},
	}
	for name, c := range cases {
		if got := From(c.content); got != c.want {
			t.Errorf("%s: From(%q) = %q, want %q", name, c.content, got, c.want)
		}
	}
}
