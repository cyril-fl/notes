// Package wikilink reads the [[links]] of a note's markdown: they are what links a note to the others.
package wikilink

import (
	"regexp"
	"slices"
	"strings"
)

var (
	link = regexp.MustCompile(`\[\[([^\[\]\n]+)\]\]`)
	// skipped is what can hold [[ ]] without being a link: fenced and inline code.
	skipped = regexp.MustCompile("(?ms)^[ \\t]*```.*?(?:^[ \\t]*```[^\\n]*$|\\z)|`[^`\\n]*`")
)

// span is one [[link]] of a text, [start, end[ with its brackets.
type span struct {
	start, end int
	target     string
}

// Parse gives the targets of content, in order, each one once whatever its case: a title, or a path
// starting with / when several notes share the title.
func Parse(content string) []string {
	var targets []string
	for _, l := range spans(content) {
		if !slices.ContainsFunc(targets, func(t string) bool { return strings.EqualFold(t, l.target) }) {
			targets = append(targets, l.target)
		}
	}
	return targets
}

// Rename writes every link to the title from as a link to the title to: [[from]], and [[/path/from]]
// whose last part is from.
func Rename(content, from, to string) string {
	links := spans(content)
	slices.Reverse(links)
	for _, l := range links {
		if renamed, ok := rename(l.target, from, to); ok {
			content = content[:l.start] + "[[" + renamed + "]]" + content[l.end:]
		}
	}
	return content
}

// Helpers

func spans(content string) []span {
	hidden := skipped.FindAllStringIndex(content, -1)
	var result []span
	for _, match := range link.FindAllStringSubmatchIndex(content, -1) {
		target := strings.TrimSpace(content[match[2]:match[3]])
		if target != "" && !inside(hidden, match[0]) {
			result = append(result, span{match[0], match[1], target})
		}
	}
	return result
}

func rename(target, from, to string) (string, bool) {
	if strings.EqualFold(target, from) {
		return to, true
	}
	if dir, last, ok := cutLast(target); ok && strings.EqualFold(last, from) {
		return dir + "/" + to, true
	}
	return "", false
}

// cutLast splits a path into its folders and its last part.
func cutLast(target string) (string, string, bool) {
	if !strings.HasPrefix(target, "/") {
		return "", "", false
	}
	i := strings.LastIndex(target, "/")
	return target[:i], target[i+1:], true
}

func inside(ranges [][]int, at int) bool {
	return slices.ContainsFunc(ranges, func(r []int) bool { return r[0] <= at && at < r[1] })
}
