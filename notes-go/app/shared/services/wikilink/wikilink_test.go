package wikilink

import (
	"slices"
	"testing"
)

func TestParse(t *testing.T) {
	cases := map[string]struct {
		content string
		want    []string
	}{
		"titles":             {"voir [[Courses]] et [[ Idées ]]", []string{"Courses", "Idées"}},
		"path":               {"[[/Projets/Idées]]", []string{"/Projets/Idées"}},
		"same whatever case": {"[[Go]] [[go]]", []string{"Go"}},
		"empty":              {"[[]] [[  ]]", nil},
		"no line break":      {"[[a\nb]]", nil},
		"inline code":        {"`[[a]]` [[b]]", []string{"b"}},
		"fenced code":        {"```\n[[a]]\n```\n[[b]]", []string{"b"}},
		"nested brackets":    {"[[[a]]]", []string{"a"}},
		"single brackets":    {"[a](b)", nil},
	}
	for name, c := range cases {
		if got := Parse(c.content); !slices.Equal(got, c.want) {
			t.Errorf("%s: Parse(%q) = %q, want %q", name, c.content, got, c.want)
		}
	}
}

func TestRename(t *testing.T) {
	cases := map[string]struct {
		content, from, to, want string
	}{
		"title":         {"voir [[Cour]] et [[cour]]", "Cour", "Courses", "voir [[Courses]] et [[Courses]]"},
		"spaces":        {"[[ Cour ]]", "Cour", "Courses", "[[Courses]]"},
		"path":          {"[[/A/Cour]]", "Cour", "Courses", "[[/A/Courses]]"},
		"other link":    {"[[Autre]] [[/Cour/x]]", "Cour", "Courses", "[[Autre]] [[/Cour/x]]"},
		"code kept":     {"`[[Cour]]` [[Cour]]", "Cour", "Courses", "`[[Cour]]` [[Courses]]"},
		"nothing to do": {"texte", "Cour", "Courses", "texte"},
	}
	for name, c := range cases {
		if got := Rename(c.content, c.from, c.to); got != c.want {
			t.Errorf("%s: Rename(%q) = %q, want %q", name, c.content, got, c.want)
		}
	}
}
