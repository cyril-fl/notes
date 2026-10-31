package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/purgeitems"
	"main.go/config"
)

var c = config.Current
var cmd = c.Commands

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
		postgrey.SetDb(env.Connect())

		schema := createschema.Process(cobra.Context(), postgrey.GetDb())
		schema.Run()
		if schema.HasErr() {
			return schema.GetErr()
		}

		// Startup is the "cron": no daemon, the trash is emptied whenever the CLI runs.
		purge := purgeitems.Process(cobra.Context(), postgrey.GetDb(), c.Trash.RetentionDays)
		purge.Run()
		return purge.GetErr()
	},
	PersistentPostRunE: func(cobra *cobra.Command, args []string) error {
		if db := postgrey.GetDb(); db != nil {
			return db.Close()
		}
		return nil
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
