// Package wikilink reads the links of a note's markdown to other notes. A link is written [[Title]] (or
// [[/Folder/Title]]) by hand; once it names exactly one note it becomes [Title](note:<id>): the id keeps
// it on its note whatever the titles do, the label keeps it readable outside the app. [[Title#Heading]]
// points to a heading of the note: [Title#Heading](note:<id>#Heading).
package wikilink

import (
	"regexp"
	"slices"
	"strings"
)

var (
	byTitle = regexp.MustCompile(`\[\[([^\[\]\n]+)\]\]`)
	byID    = regexp.MustCompile(`\[([^\[\]\n]*)\]\(note:([0-9a-fA-F-]{36})(?:#([^()\n]*))?\)`)
	// skipped is what can hold a link without being one: fenced and inline code.
	skipped = regexp.MustCompile("(?ms)^[ \\t]*```.*?(?:^[ \\t]*```[^\\n]*$|\\z)|`[^`\\n]*`")
)

// span is one link of a text, [start, end[ with its brackets: value is the target of a [[link]], the id of a note: link,
// heading the heading a note: link points to.
type span struct {
	start, end     int
	value, heading string
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

// Labels gives the label of the resolved links, by the id of their note in lowercase: the first one when several.
func Labels(content string) map[string]string {
	labels := map[string]string{}
	for _, l := range spans(content, byID, 2) {
		if id := strings.ToLower(l.value); labels[id] == "" {
			labels[id] = strings.TrimSpace(content[l.start+1 : l.start+strings.Index(content[l.start:], "](note:")])
		}
	}
	return labels
}

// HeadingLink is a resolved link to a heading of a note: [Label](note:ID#Heading).
type HeadingLink struct {
	ID, Heading, Label string
}

// HeadingLinks gives the resolved links that point to a heading, in order.
func HeadingLinks(content string) []HeadingLink {
	var links []HeadingLink
	for _, l := range spans(content, byID, 2) {
		if l.heading != "" {
			label := strings.TrimSpace(content[l.start+1 : l.start+strings.Index(content[l.start:], "](note:")])
			links = append(links, HeadingLink{strings.ToLower(l.value), l.heading, label})
		}
	}
	return links
}

// Resolve writes every [[target]] that resolve finds as a link to the note's id, labelled with its title. The
// heading after # is kept, not checked: the note's headings change with no link knowing. [[#Heading]] waits.
func Resolve(content string, resolve func(target string) (id, title string, ok bool)) string {
	links := spans(content, byTitle, 1)
	slices.Reverse(links)
	for _, l := range links {
		target, heading, _ := strings.Cut(l.value, "#")
		if target = strings.TrimSpace(target); target == "" {
			continue
		}
		if id, title, ok := resolve(target); ok {
			content = content[:l.start] + Link(id, title, strings.TrimSpace(heading)) + content[l.end:]
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
			content = content[:l.start] + Link(id, title, l.heading) + content[l.end:]
		}
	}
	return content
}

// Unresolve writes the links to notes back as [[Label]]: outside the app, the ids mean nothing. A link
// without a label stays as it is.
func Unresolve(content string) string {
	links := spans(content, byID, 2)
	slices.Reverse(links)
	for _, l := range links {
		label := content[l.start+1 : l.start+strings.Index(content[l.start:], "](note:")]
		if strings.TrimSpace(label) != "" {
			content = content[:l.start] + "[[" + label + "]]" + content[l.end:]
		}
	}
	return content
}

// Link is the markdown of a resolved link, to a heading of the note unless heading is ""; brackets would end
// the label early, parentheses the link, they are dropped.
func Link(id, title, heading string) string {
	label := strings.NewReplacer("[", "", "]", "", "\n", " ").Replace(title)
	if heading == "" {
		return "[" + label + "](note:" + id + ")"
	}
	heading = strings.NewReplacer("[", "", "]", "", "(", "", ")", "", "\n", " ").Replace(heading)
	return "[" + label + "#" + heading + "](note:" + id + "#" + heading + ")"
}

// Helpers

func spans(content string, pattern *regexp.Regexp, group int) []span {
	hidden := skipped.FindAllStringIndex(content, -1)
	var result []span
	for _, match := range pattern.FindAllStringSubmatchIndex(content, -1) {
		value := strings.TrimSpace(content[match[2*group]:match[2*group+1]])
		heading := ""
		if len(match) > 6 && match[6] >= 0 {
			heading = strings.TrimSpace(content[match[6]:match[7]])
		}
		if value != "" && !inside(hidden, match[0]) {
			result = append(result, span{match[0], match[1], value, heading})
		}
	}
	return result
}

func inside(ranges [][]int, at int) bool {
	return slices.ContainsFunc(ranges, func(r []int) bool { return r[0] <= at && at < r[1] })
}
