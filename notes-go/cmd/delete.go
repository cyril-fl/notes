package cmd

import (
	"github.com/spf13/cobra"

	cmddeletefolder "main.go/app/cli/cmd-delete-folder"
	cmddeletenote "main.go/app/cli/cmd-delete-note"
	cmddeletetag "main.go/app/cli/cmd-delete-tag"
)

var cmdDelete = &cobra.Command{
	Use:     cmd.Delete.Use,
	Short:   cmd.Delete.Short,
	Long:    cmd.Delete.Long,
	Example: cmd.Delete.Example,
}

var cmdDeleteNote = &cobra.Command{
	Use:     cmd.DeleteNote.Use,
	Short:   cmd.DeleteNote.Short,
	Long:    cmd.DeleteNote.Long,
	Example: cmd.DeleteNote.Example,
	RunE:    cmddeletenote.Delete,
}

var cmdDeleteFloder = &cobra.Command{
	Use:     cmd.DeleteFolder.Use,
	Short:   cmd.DeleteFolder.Short,
	Long:    cmd.DeleteFolder.Long,
	Example: cmd.DeleteFolder.Example,
	RunE:    cmddeletefolder.Delete,
}

var cmdDeleteTag = &cobra.Command{
	Use:     cmd.DeleteTag.Use,
	Short:   cmd.DeleteTag.Short,
	Long:    cmd.DeleteTag.Long,
	Example: cmd.DeleteTag.Example,
	RunE:    cmddeletetag.Delete,
}

func init() {
	cmdDelete.AddCommand(cmdDeleteNote, cmdDeleteFloder, cmdDeleteTag)
	cmdRoot.AddCommand(cmdDelete)
}
