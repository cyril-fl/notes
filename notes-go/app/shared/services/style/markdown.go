package style

import (
	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
)

// Markdown is glamour's standard style with the palette: no margin (aligned with the path above), text in the
// terminal color, headings and link texts in accent, rules and quotes muted.
func Markdown(isDark bool) ansi.StyleConfig {
	config := styles.LightStyleConfig
	if isDark {
		config = styles.DarkStyleConfig
	}

	accent, muted, subtle, zero := Default.Accent, Default.Muted, Default.Subtle, uint(0)
	config.Document.Margin = &zero
	config.Document.BlockPrefix = ""
	config.Document.Color = nil
	config.Heading.Color = &accent
	config.H1.Prefix, config.H1.Suffix = "", ""
	config.H1.Color, config.H1.BackgroundColor = &accent, nil
	config.H6.Color = &muted
	config.BlockQuote.Color = &muted
	config.HorizontalRule.Color = &subtle
	config.Link.Color = &muted
	config.LinkText.Color = &accent
	config.ImageText.Color = &muted
	return config
}
