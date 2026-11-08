package cmdeditnote

import (
	"context"
	"fmt"

	millflagseditnote "github.com/cyril-fl/notes/notes-go/app/cli/cmd-edit-note/services/mill"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/patchnote"
	"github.com/cyril-fl/notes/notes-go/config"
	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Edit loads the note, applies the flags (or a pre-filled form when none), then saves it.
func Edit(cobra *cobra.Command, args []string) error {
	return run(cobra.Context(), args[0], cobra.Flags())
}

// Interactive edits the note in the pre-filled form, as `edit note <id>` without flags.
func Interactive(ctx context.Context, id string) error {
	return run(ctx, id, noFlags())
}

// run edits the text only when a flag asks for it, or with the form when no flag says anything: --pin or --lock
// alone must not rewrite the note, nor change its last change date. patchnote pins, locks and edits in one transaction.
func run(ctx context.Context, id string, flags *pflag.FlagSet) error {
	input := patchnote.Input{ID: id}
	if flags.Changed("pin") || flags.Changed("unpin") {
		pinned := flags.Changed("pin")
		input.Pinned = &pinned
	}
	if flags.Changed("lock") || flags.Changed("unlock") {
		readonly := flags.Changed("lock")
		input.Readonly = &readonly
	}
	textChanged := flags.Changed("content") || flags.Changed("tag") || flags.Changed("untag") || flags.Changed("tag-from-list")
	if input.Pinned == nil && input.Readonly == nil || textChanged {
		// The mill applies --tag and --untag over the current note, patchnote then only saves its result.
		resolver := millflagseditnote.New(millflagseditnote.Template{Ctx: ctx, Db: postgrey.GetDb(), Flags: flags, ID: id})
		note := resolver.Resolve()
		if resolver.HasErr() {
			return resolver.GetErr()
		}
		input.Content, input.Tags = &note.Content, &note.Tags
	}

	process := patchnote.Process(ctx, postgrey.GetDb(), input)
	process.Run()
	if process.HasErr() {
		return process.GetErr()
	}

	if input.Content != nil {
		fmt.Printf("Successfully updated note %s\n", id)
	}
	if input.Pinned != nil && *input.Pinned {
		fmt.Printf("Successfully pinned note %s\n", id)
	} else if input.Pinned != nil {
		fmt.Printf("Successfully unpinned note %s\n", id)
	}
	if input.Readonly != nil && *input.Readonly {
		fmt.Printf("Successfully locked note %s\n", id)
	} else if input.Readonly != nil {
		fmt.Printf("Successfully unlocked note %s\n", id)
	}
	return nil
}

// noFlags declares the edit flags without setting any: the mill reads them all and sees no change.
func noFlags() *pflag.FlagSet {
	cmd := &cobra.Command{}
	plumcliutilsargs.Makeflags(cmd, config.Current.Commands.EditNote.Flags)
	return cmd.Flags()
}
