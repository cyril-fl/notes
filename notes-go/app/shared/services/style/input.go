package style

import "charm.land/bubbles/v2/textinput"

// Filter colors the prompt of a filter input with the accent.
func Filter(styles textinput.Styles) textinput.Styles {
	styles.Focused.Prompt = Accent()
	styles.Blurred.Prompt = Accent()
	return styles
}
