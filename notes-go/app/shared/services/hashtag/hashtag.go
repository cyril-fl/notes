// Package hashtag reads the #tags of a note's markdown: they are what makes the tags of the note.
package hashtag

import (
	"errors"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

var ErrInvalid = errors.New("a tag holds only letters, digits, _ or -, with at least one letter")

var (
	word = regexp.MustCompile(`#[\p{L}\p{N}_-]+`)
	// skipped is what can hold a # without being a tag: fenced code, inline code, link targets, addresses.
	skipped  = regexp.MustCompile("(?ms)^[ \\t]*```.*?(?:^[ \\t]*```[^\\n]*$|\\z)|`[^`\\n]*`|\\]\\([^)\\n]*\\)|[a-zA-Z][a-zA-Z0-9+.-]*://\\S+")
	onlyTags = regexp.MustCompile(`^[ \t]*#[\p{L}\p{N}_-]+(?:[ \t]+#[\p{L}\p{N}_-]+)*[ \t]*$`)
)

// span is one #tag of a text, [start, end[ with its #.
type span struct {
	start, end int
	value      string
}

// Valid tells whether value can be written as a #tag.
func Valid(value string) bool {
	return !strings.ContainsFunc(value, notTagRune) && strings.ContainsFunc(value, unicode.IsLetter)
}

// Parse gives the tags of content, in order, each one once whatever its case.
func Parse(content string) []string {
	var tags []string
	for _, tag := range spans(content) {
		if !contains(tags, tag.value) {
			tags = append(tags, tag.value)
		}
	}
	return tags
}

// Sync makes the #tags of content exactly tags: removes the others, adds the missing ones at the end.
func Sync(content string, tags []string) string {
	var extra, missing []string
	current := Parse(content)
	for _, tag := range current {
		if !contains(tags, tag) {
			extra = append(extra, tag)
		}
	}
	for _, tag := range tags {
		if !contains(current, tag) && !contains(missing, tag) {
			missing = append(missing, tag)
		}
	}
	return appendTags(Remove(content, extra...), missing)
}

// Rename writes every #from of content as #to.
func Rename(content, from, to string) string {
	tags := spans(content)
	slices.Reverse(tags)
	for _, tag := range tags {
		if strings.EqualFold(tag.value, from) {
			content = content[:tag.start] + "#" + to + content[tag.end:]
		}
	}
	return content
}

// Remove erases the #tags of values from content, with the space or the line they leave behind.
func Remove(content string, values ...string) string {
	tags := spans(content)
	slices.Reverse(tags)
	for _, tag := range tags {
		if contains(values, tag.value) {
			content = erase(content, tag)
		}
	}
	return strings.TrimSpace(content)
}

// Helpers

func spans(content string) []span {
	hidden := skipped.FindAllStringIndex(content, -1)
	var result []span
	for _, match := range word.FindAllStringIndex(content, -1) {
		value := content[match[0]+1 : match[1]]
		if Valid(value) && startsWord(content, match[0]) && !inside(hidden, match[0]) {
			result = append(result, span{match[0], match[1], value})
		}
	}
	return result
}

// startsWord refuses C#, ##title, \#escaped, a/#b and &#39;.
func startsWord(content string, at int) bool {
	if at == 0 {
		return true
	}
	last, _ := utf8.DecodeLastRuneInString(content[:at])
	return !unicode.IsLetter(last) && !unicode.IsDigit(last) && !strings.ContainsRune(`#\&/_-`, last)
}

func inside(ranges [][]int, at int) bool {
	return slices.ContainsFunc(ranges, func(r []int) bool { return r[0] <= at && at < r[1] })
}

func notTagRune(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-'
}

func contains(values []string, value string) bool {
	return slices.ContainsFunc(values, func(v string) bool { return strings.EqualFold(v, value) })
}

// erase drops a tag; alone on its line, the line goes too, and the blank line after it.
func erase(content string, tag span) string {
	before, after := content[:tag.start], content[tag.end:]
	lineStart := strings.TrimRight(before, " \t")
	lineEnd := strings.TrimLeft(after, " \t")

	if (lineStart == "" || strings.HasSuffix(lineStart, "\n")) && (lineEnd == "" || strings.HasPrefix(lineEnd, "\n")) {
		lineEnd = strings.TrimPrefix(lineEnd, "\n")
		return lineStart + strings.TrimPrefix(lineEnd, "\n")
	}
	if len(lineStart) < len(before) {
		return lineStart + after
	}
	return before + lineEnd
}

func appendTags(content string, tags []string) string {
	if len(tags) == 0 {
		return content
	}
	written := "#" + strings.Join(tags, " #")
	switch {
	case content == "":
		return written
	case onlyTags.MatchString(content[strings.LastIndex(content, "\n")+1:]):
		return content + " " + written
	}
	return content + "\n\n" + written
}
