package domain

import (
	"os"
	"regexp"
	"slices"
	"testing"
)

// The web app lists its themes in themes.ts: an id on one side only is refused by the API or cannot be shown.
func TestThemesMatchApp(t *testing.T) {
	file, err := os.ReadFile("../../../../notes-app/app/utils/themes.ts")
	if err != nil {
		t.Fatal(err)
	}
	var app []string
	for _, match := range regexp.MustCompile(`\{ id: '([^']+)'`).FindAllStringSubmatch(string(file), -1) {
		app = append(app, match[1])
	}
	if !slices.Equal(app, Themes) {
		t.Errorf("themes.ts has %v, domain.Themes %v", app, Themes)
	}
}
