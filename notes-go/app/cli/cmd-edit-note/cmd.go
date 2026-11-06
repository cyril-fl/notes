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
	"main.go/app/shared/services/process/pinnote"
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

// run edits the text only when a flag asks for it, or with the form when no flag says anything: --pin
// alone must not rewrite the note, nor change its last change date.
func run(ctx context.Context, id string, flags *pflag.FlagSet) error {
	pinChanged := flags.Changed("pin") || flags.Changed("unpin")
	textChanged := flags.Changed("content") || flags.Changed("tag") || flags.Changed("untag") || flags.Changed("tag-from-list")
	if pinChanged && !textChanged {
		return pin(ctx, id, flags.Changed("pin"))
	}

	resolver := millflagseditnote.New(millflagseditnote.Template{Ctx: ctx, Db: postgrey.GetDb(), Flags: flags, ID: id})
	input := resolver.Resolve()
	if resolver.HasErr() {
		return resolver.GetErr()
	}

	process := editnote.Process(ctx, postgrey.GetDb(), input)
	process.Run()
	if process.HasErr() {
		return process.GetErr()
	}
	fmt.Printf("Successfully updated note %s\n", process.GetResult().ID)

	if pinChanged {
		return pin(ctx, id, flags.Changed("pin"))
	}
	return nil
}

func pin(ctx context.Context, id string, pinned bool) error {
	process := pinnote.Process(ctx, postgrey.GetDb(), id, pinned)
	process.Run()
	if process.HasErr() {
		return process.GetErr()
	}

	if pinned {
		fmt.Printf("Successfully pinned note %s\n", id)
	} else {
		fmt.Printf("Successfully unpinned note %s\n", id)
	}
	return nil
}

// noFlags declares the edit flags without setting any: the mill reads them all and sees no change.
func noFlags() *pflag.FlagSet {
	cmd := &cobra.Command{}
	plumcliutilsargs.Makeflags(cmd, config.Current.Commands.EditNote.Flags)
	return cmd.Flags()
}
