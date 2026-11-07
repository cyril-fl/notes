package style

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// gap separates the table columns.
const gap = "  "

// Header renders the three lines above a list: bold title, muted path, subtle info.
func Header(title, path, info string) string {
	return Title().Render(title) + "\n" + Muted().Render(path) + "\n" + Subtle().Render(info) + "\n"
}

// Table aligns the cells in columns, the main one in the terminal color and the others muted. cursor marks that row
// with an accent ›; cursor < 0 marks none and drops the margin (for a pipe).
func Table(rows [][]string, cursor, main int) string {
	widths := columnWidths(rows)
	var output strings.Builder
	for i, cells := range rows {
		output.WriteString(margin(i, cursor) + line(cells, widths, main, i == cursor) + "\n")
	}
	return output.String()
}

// Helpers

func columnWidths(rows [][]string) []int {
	var widths []int
	for _, cells := range rows {
		for i, cell := range cells {
			if i == len(widths) {
				widths = append(widths, 0)
			}
			widths[i] = max(widths[i], lipgloss.Width(cell))
		}
	}
	return widths
}

func margin(row, cursor int) string {
	switch {
	case cursor < 0:
		return ""
	case row == cursor:
		return Accent().Bold(true).Render("›") + " "
	default:
		return "  "
	}
}

// line pads each cell but the last before styling it: escape codes would count in the padding.
func line(cells []string, widths []int, main int, selected bool) string {
	parts := make([]string, len(cells))
	for i, cell := range cells {
		if i < len(cells)-1 {
			cell += strings.Repeat(" ", widths[i]-lipgloss.Width(cell))
		}
		parts[i] = cellStyle(i == main, selected).Render(cell)
	}
	return strings.Join(parts, gap)
}

func cellStyle(main, selected bool) lipgloss.Style {
	switch {
	case main && selected:
		return Title()
	case main:
		return lipgloss.NewStyle()
	default:
		return Muted()
	}
}
