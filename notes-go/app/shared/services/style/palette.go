// Package style is the CLI look: one palette, named styles and the shared header and table.
package style

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Theme holds the UI colors, as hex or ANSI codes. Text has none: the terminal's own stays readable on light and dark.
type Theme struct {
	Accent  string
	Muted   string
	Subtle  string
	Error   string
	Success string
}

// Default is the only place colors live: change them here to follow the web theme.
var Default = Theme{
	Accent:  "#E5B33A",
	Muted:   "#8E8E93",
	Subtle:  "#636366",
	Error:   "#FF453A",
	Success: "#30D158",
}

func ColorAccent() color.Color  { return lipgloss.Color(Default.Accent) }
func ColorMuted() color.Color   { return lipgloss.Color(Default.Muted) }
func ColorSubtle() color.Color  { return lipgloss.Color(Default.Subtle) }
func ColorError() color.Color   { return lipgloss.Color(Default.Error) }
func ColorSuccess() color.Color { return lipgloss.Color(Default.Success) }
