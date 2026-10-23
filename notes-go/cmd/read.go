package cmd

import (
	"github.com/spf13/cobra"

	cmdreadfolder "main.go/app/cli/cmd-read-folder"
	cmdreadnote "main.go/app/cli/cmd-read-note"
	cmdreadtag "main.go/app/cli/cmd-read-tag"
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
	RunE:    cmdreadnote.Read,
}

var cmdReadFloder = &cobra.Command{
	Use:     cmd.ReadFolder.Use,
	Short:   cmd.ReadFolder.Short,
	Long:    cmd.ReadFolder.Long,
	Example: cmd.ReadFolder.Example,
	RunE:    cmdreadfolder.Read,
}

var cmdReadTag = &cobra.Command{
	Use:     cmd.ReadTag.Use,
	Short:   cmd.ReadTag.Short,
	Long:    cmd.ReadTag.Long,
	Example: cmd.ReadTag.Example,
	RunE:    cmdreadtag.Read,
}

var cmdReadFinder = &cobra.Command{
	Use:     cmd.ReadFolder.Use,
	Short:   cmd.ReadFolder.Short,
	Long:    cmd.ReadFolder.Long,
	Example: cmd.ReadFolder.Example,
	RunE:    cmdreadfolder.Read,
}

var cmdReadGraph = &cobra.Command{
	Use:     cmd.ReadTag.Use,
	Short:   cmd.ReadTag.Short,
	Long:    cmd.ReadTag.Long,
	Example: cmd.ReadTag.Example,
	RunE:    cmdreadtag.Read,
}

func init() {
	cmdRead.AddCommand(cmdReadNote, cmdReadFloder, cmdReadTag, cmdReadFinder, cmdReadGraph)
	cmdRoot.AddCommand(cmdRead)
}
