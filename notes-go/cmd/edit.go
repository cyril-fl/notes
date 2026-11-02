package cmd

import (
	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"

	cmdeditfolder "main.go/app/cli/cmd-edit-folder"
	cmdeditnote "main.go/app/cli/cmd-edit-note"
	cmdedittag "main.go/app/cli/cmd-edit-tag"
	"main.go/app/shared/domain"
)

var cmdEdit = &cobra.Command{
	Use:     cmd.Edit.Use,
	Short:   cmd.Edit.Short,
	Long:    cmd.Edit.Long,
	Example: cmd.Edit.Example,
}

var cmdEditNote = &cobra.Command{
	Use:     cmd.EditNote.Use,
	Short:   cmd.EditNote.Short,
	Long:    cmd.EditNote.Long,
	Example: cmd.EditNote.Example,
	Args:    cobra.ExactArgs(1),
	RunE:    resolved(domain.NOTE, cmdeditnote.Edit),
}

var cmdEditFloder = &cobra.Command{
	Use:     cmd.EditFolder.Use,
	Short:   cmd.EditFolder.Short,
	Long:    cmd.EditFolder.Long,
	Example: cmd.EditFolder.Example,
	Args:    cobra.ExactArgs(1),
	RunE:    resolved(domain.FOLDER, cmdeditfolder.Edit),
}

var cmdEditTag = &cobra.Command{
	Use:     cmd.EditTag.Use,
	Short:   cmd.EditTag.Short,
	Long:    cmd.EditTag.Long,
	Example: cmd.EditTag.Example,
	Args:    cobra.ExactArgs(2),
	RunE:    cmdedittag.Edit,
}

func init() {
	plumcliutilsargs.Makeflags(cmdEditNote, cmd.EditNote.Flags)
	plumcliutilsargs.Makeflags(cmdEditFloder, cmd.EditFolder.Flags)

	cmdEdit.AddCommand(cmdEditNote, cmdEditFloder, cmdEditTag)
	cmdRoot.AddCommand(cmdEdit)
}
