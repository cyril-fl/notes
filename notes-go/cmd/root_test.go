package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNeedsDb(t *testing.T) {
	root := &cobra.Command{Use: "notes"}
	show := &cobra.Command{Use: "show"}
	note := &cobra.Command{Use: "note"}
	completion := &cobra.Command{Use: "completion"}
	zsh := &cobra.Command{Use: "zsh"}
	request := &cobra.Command{Use: cobra.ShellCompRequestCmd}
	show.AddCommand(note)
	completion.AddCommand(zsh)
	root.AddCommand(show, completion, request)

	tests := map[string]struct {
		command *cobra.Command
		want    bool
	}{
		"root help":        {root, false},
		"subcommand":       {note, true},
		"completion":       {completion, false},
		"completion zsh":   {zsh, false},
		"shell __complete": {request, false},
	}

	for name, test := range tests {
		if got := needsDb(test.command); got != test.want {
			t.Errorf("%s: needsDb() = %v, want %v", name, got, test.want)
		}
	}
}
