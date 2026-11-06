// Package wikilink reads the links of a note's markdown to other notes. A link is written [[Title]] (or
// [[/Folder/Title]]) by hand; once it names exactly one note it becomes [Title](note:<id>): the id keeps
// it on its note whatever the titles do, the label keeps it readable outside the app.
package wikilink

import (
	"regexp"
	"slices"
	"strings"
)

var (
	byTitle = regexp.MustCompile(`\[\[([^\[\]\n]+)\]\]`)
	byID    = regexp.MustCompile(`\[([^\[\]\n]*)\]\(note:([0-9a-fA-F-]{36})\)`)
	// skipped is what can hold a link without being one: fenced and inline code.
	skipped = regexp.MustCompile("(?ms)^[ \\t]*```.*?(?:^[ \\t]*```[^\\n]*$|\\z)|`[^`\\n]*`")
)

// span is one link of a text, [start, end[ with its brackets: value is the target of a [[link]], the id of a note: link.
type span struct {
	start, end int
	value      string
}

// Parse gives the targets of the [[links]] not resolved yet, in order, each one once whatever its case.
func Parse(content string) []string {
	var targets []string
	for _, l := range spans(content, byTitle, 1) {
		if !slices.ContainsFunc(targets, func(t string) bool { return strings.EqualFold(t, l.value) }) {
			targets = append(targets, l.value)
		}
	}
	return targets
}

// IDs gives the ids of the notes the resolved links point to, in order, each one once.
func IDs(content string) []string {
	var ids []string
	for _, l := range spans(content, byID, 2) {
		if id := strings.ToLower(l.value); !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	return ids
}

// Resolve writes every [[target]] that resolve finds as a link to the note's id, labelled with its title.
func Resolve(content string, resolve func(target string) (id, title string, ok bool)) string {
	links := spans(content, byTitle, 1)
	slices.Reverse(links)
	for _, l := range links {
		if id, title, ok := resolve(l.value); ok {
			content = content[:l.start] + Link(id, title) + content[l.end:]
		}
	}
	return content
}

// Relabel gives the links to the note id the label title: the text says what the note is called now.
func Relabel(content, id, title string) string {
	links := spans(content, byID, 2)
	slices.Reverse(links)
	for _, l := range links {
		if strings.EqualFold(l.value, id) {
			content = content[:l.start] + Link(id, title) + content[l.end:]
		}
	}
	return content
}

// Link is the markdown of a resolved link; brackets would end the label early, they are dropped.
func Link(id, title string) string {
	label := strings.NewReplacer("[", "", "]", "", "\n", " ").Replace(title)
	return "[" + label + "](note:" + id + ")"
}

// Helpers

func spans(content string, pattern *regexp.Regexp, group int) []span {
	hidden := skipped.FindAllStringIndex(content, -1)
	var result []span
	for _, match := range pattern.FindAllStringSubmatchIndex(content, -1) {
		value := strings.TrimSpace(content[match[2*group]:match[2*group+1]])
		if value != "" && !inside(hidden, match[0]) {
			result = append(result, span{match[0], match[1], value})
		}
	}
	return result
}

func inside(ranges [][]int, at int) bool {
	return slices.ContainsFunc(ranges, func(r []int) bool { return r[0] <= at && at < r[1] })
}
