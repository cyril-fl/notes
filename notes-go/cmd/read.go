package cmd

import (
	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"

	cmdreadfolder "github.com/cyril-fl/notes/notes-go/app/cli/cmd-read-folder"
	cmdreadnote "github.com/cyril-fl/notes/notes-go/app/cli/cmd-read-note"
	cmdreadtag "github.com/cyril-fl/notes/notes-go/app/cli/cmd-read-tag"
	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
)

var cmdRead = &cobra.Command{
	Use:     cmd.Read.Use,
	Short:   cmd.Read.Short,
	Long:    cmd.Read.Long,
	Example: cmd.Read.Example,
}

var cmdReadNote = &cobra.Command{
	Use:     cmd.ReadNote.Use,
	Short:   cmd.ReadNote.Short,
	Long:    cmd.ReadNote.Long,
	Example: cmd.ReadNote.Example,
	Args:    cobra.ExactArgs(1),
	RunE:    resolved(domain.NOTE, cmdreadnote.Read),
}

var cmdReadFloder = &cobra.Command{
	Use:     cmd.ReadFolder.Use,
	Short:   cmd.ReadFolder.Short,
	Long:    cmd.ReadFolder.Long,
	Example: cmd.ReadFolder.Example,
	Args:    cobra.ExactArgs(1),
	RunE:    resolved(domain.FOLDER, cmdreadfolder.Read),
}

var cmdReadTag = &cobra.Command{
	Use:     cmd.ReadTag.Use,
	Short:   cmd.ReadTag.Short,
	Long:    cmd.ReadTag.Long,
	Example: cmd.ReadTag.Example,
	Args:    cobra.MaximumNArgs(1),
	RunE:    cmdreadtag.Read,
}

func init() {
	plumcliutilsargs.Makeflags(cmdReadTag, cmd.ReadTag.Flags)
	cmdReadTag.MarkFlagsMutuallyExclusive("yaml", "json")
	plumcliutilsargs.Makeflags(cmdReadFloder, cmd.ReadFolder.Flags)

	cmdRead.AddCommand(cmdReadNote, cmdReadFloder, cmdReadTag)
	cmdRoot.AddCommand(cmdRead)
}
