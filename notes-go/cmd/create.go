package cmd

import (
	"github.com/spf13/cobra"
	src "main.go/app/cmdcreate"
)

/*
TODO Add suub commands for create command
var cmdCreateNote = &cobra.Command{
	Use:     cmd.cmdCreateNote.Use,
	Short:   cmd.cmdCreateNote.Short,
	Long:    cmd.cmdCreateNote.Long,
	Example: cmd.cmdCreateNote.Example,
	RunE:    src.cmdCreateNote,
}

var cmdCreateFloder = &cobra.Command{
	Use:     cmd.cmdCreateFloder.Use,
	Short:   cmd.cmdCreateFloder.Short,
	Long:    cmd.cmdCreateFloder.Long,
	Example: cmd.cmdCreateFloder.Example,
	RunE:    src.cmdCreateFloder,
}

var cmdCreateTag = &cobra.Command{
	Use:     cmd.cmdCreateTag.Use,
	Short:   cmd.cmdCreateTag.Short,
	Long:    cmd.cmdCreateTag.Long,
	Example: cmd.cmdCreateTag.Example,
	RunE:    src.cmdCreateTag,
}

cmdCreate.AddCommand(cmdCreateNote, cmdCreateFloder, cmdCreateTag)
*/

var cmdCreate = &cobra.Command{
	Use:     cmd.Create.Use,
	Short:   cmd.Create.Short,
	Long:    cmd.Create.Long,
	Example: cmd.Create.Example,
	RunE:    src.Create,
}

func init() {
	cmdRoot.AddCommand(cmdCreate)
}
