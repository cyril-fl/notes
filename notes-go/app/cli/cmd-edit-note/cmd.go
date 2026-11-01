package cmdeditnote

import (
	"context"
	"fmt"

	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	millflagseditnote "main.go/app/cli/cmd-edit-note/services/mill"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/editnote"
	"main.go/config"
)

// Edit loads the note, applies the flags (or a pre-filled form when none), then saves it.
func Edit(cobra *cobra.Command, args []string) error {
	return run(cobra.Context(), args[0], cobra.Flags())
}

// Interactive edits the note in the pre-filled form, as `edit note <id>` without flags.
func Interactive(ctx context.Context, id string) error {
	return run(ctx, id, noFlags())
}

func run(ctx context.Context, id string, flags *pflag.FlagSet) error {
	resolver := millflagseditnote.New(millflagseditnote.Template{Ctx: ctx, Db: postgrey.GetDb(), Flags: flags, ID: id})
	input := resolver.Resolve()
	if resolver.HasErr() {
		return resolver.GetErr()
	}

	process := editnote.Process(ctx, postgrey.GetDb(), input)
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully updated note %s\n", process.GetResult().ID)
	}

	return process.GetErr()
}

// noFlags declares the edit flags without setting any: the mill reads them all and sees no change.
func noFlags() *pflag.FlagSet {
	cmd := &cobra.Command{}
	plumcliutilsargs.Makeflags(cmd, config.Current.Commands.EditNote.Flags)
	return cmd.Flags()
}
