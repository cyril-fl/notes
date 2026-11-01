package cmdeditfolder

import (
	"context"
	"fmt"

	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	millflagseditfolder "main.go/app/cli/cmd-edit-folder/services/mill"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/editfolder"
	"main.go/config"
)

// Edit loads the folder, applies --title (or a pre-filled form), then saves it.
func Edit(cobra *cobra.Command, args []string) error {
	return run(cobra.Context(), args[0], cobra.Flags())
}

// Interactive edits the folder in the pre-filled form, as `edit folder <id>` without flags.
func Interactive(ctx context.Context, id string) error {
	return run(ctx, id, noFlags())
}

func run(ctx context.Context, id string, flags *pflag.FlagSet) error {
	resolver := millflagseditfolder.New(millflagseditfolder.Template{Ctx: ctx, Db: postgrey.GetDb(), Flags: flags, ID: id})
	input := resolver.Resolve()
	if resolver.HasErr() {
		return resolver.GetErr()
	}

	process := editfolder.Process(ctx, postgrey.GetDb(), input)
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully updated folder %s\n", process.GetResult().ID)
	}

	return process.GetErr()
}

// noFlags declares the edit flags without setting any: the mill sees no --title and opens the form.
func noFlags() *pflag.FlagSet {
	cmd := &cobra.Command{}
	plumcliutilsargs.Makeflags(cmd, config.Current.Commands.EditFolder.Flags)
	return cmd.Flags()
}
