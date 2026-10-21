package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/uptrace/bun"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createschema"
	"main.go/config"
)

var c = config.Current
var cmd = c.Commands
var db *bun.DB

// TODO appliquer ca a solopg
var cmdRoot = &cobra.Command{
	Use:   c.Name,
	Short: cmd.Root.Short,
	Long:  cmd.Root.Long,
	/* Version: config.Version, 	// TODO link ca correctement */
	Example: cmd.Root.Example,

	// Runs before every subcommand: schema creation is idempotent (IF NOT EXISTS).
	PersistentPreRunE: func(cobra *cobra.Command, args []string) error {
		env := postgrey.NewEnvironment()
		env.ReadEnv()
		db = env.Connect()

		process := createschema.Process(cobra.Context(), db)
		process.Run()
		return process.GetErr()
	},
	PersistentPostRunE: func(cobra *cobra.Command, args []string) error {
		return db.Close()
	},
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
