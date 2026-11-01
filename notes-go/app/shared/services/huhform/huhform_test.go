package huhform

import (
	"testing"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"main.go/app/shared/services/process/listfolders"
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

func TestFolderOptions(t *testing.T) {
	options := folderOptions([]listfolders.Folder{{ID: "a", Path: "/A"}, {ID: "b", Path: "/A/B"}})
	if len(options) != 3 || options[0].Key != "/" || options[0].Value != "" || options[2].Key != "/A/B" || options[2].Value != "b" {
		t.Errorf("options = %+v, want / then the paths", options)
	}
}
