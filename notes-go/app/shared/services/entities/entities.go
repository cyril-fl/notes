// Package entities undoes the HTML entities the web editor once saved for & < > in text (issue #1).
package entities

import (
	"regexp"
	"strings"
)

var (
	entity = strings.NewReplacer("&lt;", "<", "&gt;", ">", "&amp;", "&")
	// code is what the editor writes as typed: fenced and inline code, where an &lt; is the text itself.
	code = regexp.MustCompile("(?ms)^[ \\t]*(?:```|~~~).*?(?:^[ \\t]*(?:```|~~~)[^\\n]*$|\\z)|`[^`\\n]*`")
)

// Decode turns &lt; &gt; &amp; back into < > & outside code, in one pass: "&amp;lt;" gives "&lt;".
func Decode(content string) string {
	var out strings.Builder
	last := 0
	for _, span := range code.FindAllStringIndex(content, -1) {
		out.WriteString(entity.Replace(content[last:span[0]]))
		out.WriteString(content[span[0]:span[1]])
		last = span[1]
	}
	out.WriteString(entity.Replace(content[last:]))
	return out.String()
}
