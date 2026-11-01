package huhform

import (
	"testing"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

func TestFitFillsHeight(t *testing.T) {
	var title, content string
	tests := map[string][]huh.Field{
		"text grows": {huh.NewInput().Title("Title").Value(&title), huh.NewText().Title("Content").Value(&content)},
		"form pads":  {huh.NewInput().Title("Title").Value(&title)},
	}

	for name, fields := range tests {
		form := fit(fields, 80, 30)
		form.Init()
		if got := lipgloss.Height(form.View()); got != 30 {
			t.Errorf("%s: height = %d, want 30", name, got)
		}
	}
}
