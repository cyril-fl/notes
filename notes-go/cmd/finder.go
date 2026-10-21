package cmd

import (
	"github.com/spf13/cobra"
	src "main.go/app/cmdfinder"
)

var cmdFinder = &cobra.Command{
	Use:     cmd.Finder.Use,
	Short:   cmd.Finder.Short,
	Long:    cmd.Finder.Long,
	Example: cmd.Finder.Example,
	RunE:    src.Finder,
}

func init() {
	cmdRoot.AddCommand(cmdFinder)
}
