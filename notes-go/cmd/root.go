package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"main.go/config"
)

var c = config.Current
var cmd = c.Commands

var cmdRoot = &cobra.Command{
	Use:   c.Name,
	Short: cmd.Root.Short,
	Long:  cmd.Root.Long,
	// TODO link ca correctement
	// Version: config.Version,
	Example: cmd.Root.Example,

	RunE: func(cobra *cobra.Command, args []string) error {
		return cobra.Help()
	},
}

func Execute() {
	if err := cmdRoot.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
