// Package notetitle reads the title of a note from its markdown, as Apple Notes does: nobody types it apart.
package notetitle

import (
	"regexp"
	"strings"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/hashtag"
)

// MaxLength is the most runes a title keeps when it comes from the first words of the text.
const MaxLength = 50

var (
	fence   = regexp.MustCompile("^[ \\t]*(```|~~~)")
	heading = regexp.MustCompile(`^ {0,3}#{1,6}(?:[ \t]+(.*?))?(?:[ \t]+#+)?[ \t]*$`)
	// block is what starts a line without being text: quote, list item, task box.
	block = regexp.MustCompile(`^[ \t]*(?:>[ \t]?)*(?:(?:[-*+]|\d+[.)])[ \t]+)?(?:\[[ xX]\][ \t]+)?`)
	image = regexp.MustCompile(`!\[([^\]]*)\]\([^)]*\)`)
	link  = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	wiki  = regexp.MustCompile(`\[\[([^\[\]\n]+)\]\]`)
	marks = regexp.MustCompile("\\*\\*|__|~~|[*_`]")
	rule  = regexp.MustCompile(`^[ \t]*(?:[-*_][ \t]*){3,}$`)
)

// From gives the first heading of content, else its first line of text cut to MaxLength runes;
// "" when there is neither: the note is untitled. Code blocks and lines of #tags only are skipped.
func From(content string) string {
	if title := Heading(content); title != "" {
		return title
	}
	lines := textLines(content)
	for _, line := range lines {
		if heading.MatchString(line) || rule.MatchString(line) {
			continue
		}
		if title := clean(block.ReplaceAllString(line, "")); title != "" {
			return cut(title)
		}
	}
	return ""
}

// Heading gives the first heading of content, "" without one.
func Heading(content string) string {
	if headings := Headings(content); len(headings) > 0 {
		return headings[0]
	}
	return ""
}

// Headings gives the headings of content, in order, as their words: those of fenced code and empty ones left out.
func Headings(content string) []string {
	var headings []string
	for _, line := range textLines(content) {
		if match := heading.FindStringSubmatch(line); match != nil {
			if title := clean(match[1]); title != "" {
				headings = append(headings, title)
			}
		}
	}
	return headings
}

// Helpers

// textLines drops fenced code: a # there is a comment, not a heading.
func textLines(content string) []string {
	var lines []string
	inCode := false
	for line := range strings.Lines(content) {
		if fence.MatchString(line) {
			inCode = !inCode
			continue
		}
		if !inCode {
			lines = append(lines, strings.TrimRight(line, "\r\n"))
		}
	}
	return lines
}

// clean keeps the words: no link targets, [[ ]] around note links, emphasis marks or #tags.
func clean(text string) string {
	text = wiki.ReplaceAllString(text, "$1")
	text = image.ReplaceAllString(text, "$1")
	text = link.ReplaceAllString(text, "$1")
	text = marks.ReplaceAllString(text, "")
	text = hashtag.Remove(text, hashtag.Parse(text)...)
	return strings.Join(strings.Fields(text), " ")
}

func cut(title string) string {
	runes := []rune(title)
	if len(runes) <= MaxLength {
		return title
	}
	return strings.TrimSpace(string(runes[:MaxLength]))
}
