package style

import "charm.land/lipgloss/v2"

func Title() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true)
}

func Accent() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ColorAccent())
}

func Muted() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ColorMuted())
}

func Subtle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ColorSubtle())
}

func Error() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ColorError())
}

// Key renders a shortcut in the footer.
func Key() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(ColorAccent())
}

// Empty renders the message of a list without rows.
func Empty() lipgloss.Style {
	return Muted().Italic(true)
}
