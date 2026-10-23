package cmd

import (
	"github.com/spf13/cobra"

	cmdeditfolder "main.go/app/cli/cmd-edit-folder"
	cmdeditnote "main.go/app/cli/cmd-edit-note"
	cmdedittag "main.go/app/cli/cmd-edit-tag"
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
	RunE:    cmdeditnote.Edit,
}

var cmdEditFloder = &cobra.Command{
	Use:     cmd.EditFolder.Use,
	Short:   cmd.EditFolder.Short,
	Long:    cmd.EditFolder.Long,
	Example: cmd.EditFolder.Example,
	RunE:    cmdeditfolder.Edit,
}

var cmdEditTag = &cobra.Command{
	Use:     cmd.EditTag.Use,
	Short:   cmd.EditTag.Short,
	Long:    cmd.EditTag.Long,
	Example: cmd.EditTag.Example,
	RunE:    cmdedittag.Edit,
}

func init() {
	cmdEdit.AddCommand(cmdEditNote, cmdEditFloder, cmdEditTag)
	cmdRoot.AddCommand(cmdEdit)
}
