package cmdcreatenote

import (
	"context"
	"fmt"

	millflagscreatenote "github.com/cyril-fl/notes/notes-go/app/cli/cmd-create-note/services/mill"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/config"
	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Create resolves the note from flags (and a form for what is missing), then saves it.
func Create(cobra *cobra.Command, args []string) error {
	return run(cobra.Context(), cobra.Flags())
}

// Interactive creates a note in the form, inside parentID ("" = root).
func Interactive(ctx context.Context, parentID string) error {
	flags, err := parentFlags(parentID)
	if err != nil {
		return err
	}
	return run(ctx, flags)
}

func run(ctx context.Context, flags *pflag.FlagSet) error {
	resolver := millflagscreatenote.New(millflagscreatenote.Template{Ctx: ctx, Db: postgrey.GetDb(), Flags: flags})
	input := resolver.Resolve()
	if resolver.HasErr() {
		return resolver.GetErr()
	}

	process := createnote.Process(ctx, postgrey.GetDb(), input)
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully created note %s\n", process.GetResult().ID)
	}

	return process.GetErr()
}

// parentFlags declares the add flags with only --parent set: the mill asks for the rest.
func parentFlags(parentID string) (*pflag.FlagSet, error) {
	cmd := &cobra.Command{}
	plumcliutilsargs.Makeflags(cmd, config.Current.Commands.CreateNote.Flags)
	if parentID == "" {
		return cmd.Flags(), nil
	}
	return cmd.Flags(), cmd.Flags().Set("parent", parentID)
}
