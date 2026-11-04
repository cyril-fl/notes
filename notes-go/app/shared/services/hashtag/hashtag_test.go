package hashtag

import (
	"slices"
	"testing"
)

func TestParse(t *testing.T) {
	cases := map[string]struct {
		content string
		want    []string
	}{
		"words":              {"du #go et #idée, puis (#a_b) **#go-lang**", []string{"go", "idée", "a_b", "go-lang"}},
		"start of line":      {"#go\n#sql", []string{"go", "sql"}},
		"same whatever case": {"#Go #go #GO", []string{"Go"}},
		"titles":             {"# Titre\n## Autre\n##deux", nil},
		"numbers":            {"#123 issue", nil},
		"escaped":            {`\#go`, nil},
		"inside a word":      {"C# a#b &#39;", nil},
		"inline code":        {"`#go` #sql", []string{"sql"}},
		"fenced code":        {"```c\n#include <a>\n```\n#sql", []string{"sql"}},
		"unclosed fence":     {"```\n#go", nil},
		"link target":        {"[x](http://a.b/#go) [y](/p#sql)", nil},
		"link text":          {"[#go](http://a.b)", []string{"go"}},
		"address":            {"http://a.b/c#go fin", nil},
	}
	for name, c := range cases {
		if got := Parse(c.content); !slices.Equal(got, c.want) {
			t.Errorf("%s: Parse(%q) = %q, want %q", name, c.content, got, c.want)
		}
	}
}

func TestValid(t *testing.T) {
	for value, want := range map[string]bool{"go": true, "idée": true, "go-lang_2": true, "123": false, "a b": false, "#go": false, "": false} {
		if got := Valid(value); got != want {
			t.Errorf("Valid(%q) = %v", value, got)
		}
	}
}

func TestSync(t *testing.T) {
	cases := map[string]struct {
		content string
		tags    []string
		want    string
	}{
		"nothing to do":    {"du #go", []string{"GO"}, "du #go"},
		"adds a paragraph": {"texte", []string{"go", "sql"}, "texte\n\n#go #sql"},
		"adds to tag line": {"texte\n\n#go", []string{"go", "sql"}, "texte\n\n#go #sql"},
		"adds to empty":    {"", []string{"go"}, "#go"},
		"removes":          {"du #go et #sql", []string{"sql"}, "du et #sql"},
		"removes and adds": {"du #go", []string{"sql"}, "du\n\n#sql"},
		"keeps the code":   {"`#go` #sql", nil, "`#go`"},
	}
	for name, c := range cases {
		if got := Sync(c.content, c.tags); got != c.want {
			t.Errorf("%s: Sync(%q, %q) = %q, want %q", name, c.content, c.tags, got, c.want)
		}
	}
}

func TestRemove(t *testing.T) {
	cases := map[string]struct{ content, want string }{
		"middle":        {"a #go b", "a b"},
		"start":         {"#go a", "a"},
		"end":           {"a #go", "a"},
		"every case":    {"#GO a #go", "a"},
		"alone on line": {"a\n\n#go\n\nb", "a\n\nb"},
		"only the tag":  {"#go", ""},
		"other kept":    {"a #golang #go-lang", "a #golang #go-lang"},
		"code kept":     {"```\n#go\n```\n\n#go", "```\n#go\n```"},
		"lines kept":    {"a\n\n\nb #go", "a\n\n\nb"},
	}
	for name, c := range cases {
		if got := Remove(c.content, "go"); got != c.want {
			t.Errorf("%s: Remove(%q) = %q, want %q", name, c.content, got, c.want)
		}
	}
}

func TestRename(t *testing.T) {
	got := Rename("#Go, #golang et `#go` #go", "go", "lang")
	if want := "#lang, #golang et `#go` #lang"; got != want {
		t.Errorf("Rename() = %q, want %q", got, want)
	}
}
