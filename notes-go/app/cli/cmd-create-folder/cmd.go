package cmdcreatefolder

import (
	"context"
	"fmt"

	millflagscreatefolder "github.com/cyril-fl/notes/notes-go/app/cli/cmd-create-folder/services/mill"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createfolder"
	"github.com/cyril-fl/notes/notes-go/config"
	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Create resolves the folder from flags (and a form for what is missing), then saves it.
func Create(cobra *cobra.Command, args []string) error {
	return run(cobra.Context(), cobra.Flags())
}

// Interactive creates a folder in the form, inside parentID ("" = root).
func Interactive(ctx context.Context, parentID string) error {
	flags, err := parentFlags(parentID)
	if err != nil {
		return err
	}
	return run(ctx, flags)
}

func run(ctx context.Context, flags *pflag.FlagSet) error {
	resolver := millflagscreatefolder.New(millflagscreatefolder.Template{Ctx: ctx, Db: postgrey.GetDb(), Flags: flags})
	input := resolver.Resolve()
	if resolver.HasErr() {
		return resolver.GetErr()
	}

	process := createfolder.Process(ctx, postgrey.GetDb(), input)
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully created folder %s\n", process.GetResult().ID)
	}

	return process.GetErr()
}

// parentFlags declares the add flags with only --parent set: the mill asks for the rest.
func parentFlags(parentID string) (*pflag.FlagSet, error) {
	cmd := &cobra.Command{}
	plumcliutilsargs.Makeflags(cmd, config.Current.Commands.CreateFolder.Flags)
	if parentID == "" {
		return cmd.Flags(), nil
	}
	return cmd.Flags(), cmd.Flags().Set("parent", parentID)
}
