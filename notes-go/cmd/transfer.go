package cmd

import (
	cmdexport "github.com/cyril-fl/notes/notes-go/app/cli/cmd-export"
	cmdimport "github.com/cyril-fl/notes/notes-go/app/cli/cmd-import"
	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
)

var cmdExport = &cobra.Command{
	Use:     cmd.Export.Use,
	Short:   cmd.Export.Short,
	Long:    cmd.Export.Long,
	Example: cmd.Export.Example,
	Args:    cobra.RangeArgs(1, 2),
	RunE:    cmdexport.Export,
}

var cmdImport = &cobra.Command{
	Use:     cmd.Import.Use,
	Short:   cmd.Import.Short,
	Long:    cmd.Import.Long,
	Example: cmd.Import.Example,
	Args:    cobra.ExactArgs(1),
	RunE:    cmdimport.Import,
}

func init() {
	plumcliutilsargs.Makeflags(cmdImport, cmd.Import.Flags)

	cmdRoot.AddCommand(cmdExport, cmdImport)
}
