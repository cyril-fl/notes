package wikilink

import (
	"maps"
	"slices"
	"strings"
	"testing"
)

const (
	idA = "11111111-1111-1111-1111-111111111111"
	idB = "22222222-2222-2222-2222-222222222222"
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
		"resolved":           {"[a](note:" + idA + ")", nil},
		"heading":            {"[a#b](note:" + idA + "#b)", nil},
	}
	for name, c := range cases {
		if got := Parse(c.content); !slices.Equal(got, c.want) {
			t.Errorf("%s: Parse(%q) = %q, want %q", name, c.content, got, c.want)
		}
	}
}

func TestIDs(t *testing.T) {
	cases := map[string]struct {
		content string
		want    []string
	}{
		"links":       {"[A](note:" + idA + ") puis [B](note:" + idB + ")", []string{idA, idB}},
		"once":        {"[A](note:" + idA + ") [a](note:" + idA + ")", []string{idA}},
		"empty label": {"[](note:" + idA + ")", []string{idA}},
		"code":        {"`[A](note:" + idA + ")`", nil},
		"web link":    {"[A](https://a.b)", nil},
		"not an id":   {"[A](note:nope)", nil},
		"by title":    {"[[A]]", nil},
		"heading":     {"[A#b](note:" + idA + "#b c)", []string{idA}},
	}
	for name, c := range cases {
		if got := IDs(c.content); !slices.Equal(got, c.want) {
			t.Errorf("%s: IDs(%q) = %q, want %q", name, c.content, got, c.want)
		}
	}
}

func TestLabels(t *testing.T) {
	got := Labels("[A](note:" + idA + ") [a](note:" + strings.ToUpper(idA) + ") [B#h](note:" + idB + "#h) `[C](note:" + idB + ")` [[D]]")
	want := map[string]string{idA: "A", idB: "B#h"}
	if !maps.Equal(got, want) {
		t.Errorf("Labels = %q, want %q", got, want)
	}
}

func TestResolve(t *testing.T) {
	notes := map[string][2]string{"courses": {idA, "Courses"}, "/a/b": {idB, "B"}}
	resolve := func(target string) (string, string, bool) {
		note, ok := notes[strings.ToLower(target)]
		return note[0], note[1], ok
	}
	got := Resolve("[[courses]], [[/a/b]], [[inconnue]] et `[[courses]]`, [[courses # Lait (bio)]] [[#Haut]]", resolve)
	want := "[Courses](note:" + idA + "), [B](note:" + idB + "), [[inconnue]] et `[[courses]]`, [Courses#Lait bio](note:" + idA + "#Lait bio) [[#Haut]]"
	if got != want {
		t.Errorf("Resolve = %q, want %q", got, want)
	}
}

func TestRelabel(t *testing.T) {
	got := Relabel("[Cour](note:"+idA+") [B](note:"+idB+") [Cour#Lait](note:"+idA+"#Lait)", idA, "Courses [x]")
	if want := "[Courses x](note:" + idA + ") [B](note:" + idB + ") [Courses x#Lait](note:" + idA + "#Lait)"; got != want {
		t.Errorf("Relabel = %q, want %q", got, want)
	}
}

func TestUnresolve(t *testing.T) {
	got := Unresolve("[Courses](note:" + idA + ") [](note:" + idB + ") `[A](note:" + idA + ")` [[B]] [Courses#Lait](note:" + idA + "#Lait)")
	if want := "[[Courses]] [](note:" + idB + ") `[A](note:" + idA + ")` [[B]] [[Courses#Lait]]"; got != want {
		t.Errorf("Unresolve = %q, want %q", got, want)
	}
}

// An embed ![[Note]] is a link after a !: resolved to ![Note](note:<id>), given back as ![[Note]] on export.
func TestEmbed(t *testing.T) {
	resolve := func(string) (string, string, bool) { return idA, "Courses", true }
	resolved := Resolve("![[courses]]", resolve)
	if want := "![Courses](note:" + idA + ")"; resolved != want {
		t.Errorf("Resolve = %q, want %q", resolved, want)
	}
	if got := Unresolve(resolved); got != "![[Courses]]" {
		t.Errorf("Unresolve = %q, want %q", got, "![[Courses]]")
	}
}
