package cmd

import (
	"github.com/spf13/cobra"

	cmdcreatefolder "main.go/app/cli/cmd-create-folder"
	cmdcreatenote "main.go/app/cli/cmd-create-note"
	cmdcreatetag "main.go/app/cli/cmd-create-tag"
)

var cmdCreate = &cobra.Command{
	Use:     cmd.Create.Use,
	Short:   cmd.Create.Short,
	Long:    cmd.Create.Long,
	Example: cmd.Create.Example,
}

var cmdCreateNote = &cobra.Command{
	Use:     cmd.CreateNote.Use,
	Short:   cmd.CreateNote.Short,
	Long:    cmd.CreateNote.Long,
	Example: cmd.CreateNote.Example,
	RunE:    cmdcreatenote.Create,
}

var cmdCreateFloder = &cobra.Command{
	Use:     cmd.CreateFolder.Use,
	Short:   cmd.CreateFolder.Short,
	Long:    cmd.CreateFolder.Long,
	Example: cmd.CreateFolder.Example,
	RunE:    cmdcreatefolder.Create,
}

var cmdCreateTag = &cobra.Command{
	Use:     cmd.CreateTag.Use,
	Short:   cmd.CreateTag.Short,
	Long:    cmd.CreateTag.Long,
	Example: cmd.CreateTag.Example,
	RunE:    cmdcreatetag.Create,
}

func init() {
	cmdCreate.AddCommand(cmdCreateNote, cmdCreateFloder, cmdCreateTag)
	cmdRoot.AddCommand(cmdCreate)
}
