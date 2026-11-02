package config

import "testing"

// Outside the repo there is no CONFIG_FILE: the config built into the binary must load.
func TestLoadEmbedded(t *testing.T) {
	t.Setenv("CONFIG_FILE", "")

	loaded, err := load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Name != "notes" || loaded.Commands.Search.Use == "" {
		t.Errorf("loaded = %q, search %q", loaded.Name, loaded.Commands.Search.Use)
	}
}

func TestLoadFile(t *testing.T) {
	t.Setenv("CONFIG_FILE", "config.yaml")

	loaded, err := load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Name != "notes" {
		t.Errorf("name = %q", loaded.Name)
	}
}
