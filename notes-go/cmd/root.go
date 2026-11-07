package cmd

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
	cmdaccount "main.go/app/cli/cmd-account"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/purgeitems"
	"main.go/app/shared/services/screen"
	"main.go/config"
)

var c = config.Current
var cmd = c.Commands

// TODO appliquer ca a solopg
var cmdRoot = &cobra.Command{
	Use:     c.Name,
	Short:   cmd.Root.Short,
	Long:    cmd.Root.Long,
	Version: version(),
	Example: cmd.Root.Example,
	// Errors are printed by Execute so a form abort (Ctrl+C / Esc) stays silent.
	SilenceUsage:  true,
	SilenceErrors: true,

	// Runs before every subcommand: schema creation is idempotent (IF NOT EXISTS).
	PersistentPreRunE: func(cobra *cobra.Command, args []string) error {
		if !needsDb(cobra) {
			return nil
		}

		env := postgrey.NewEnvironment()
		env.ReadEnv()
		postgrey.SetDb(env.Connect())

		schema := createschema.Process(cobra.Context(), postgrey.GetDb())
		schema.Run()
		if schema.HasErr() {
			return schema.GetErr()
		}

		// Startup is the "cron": no daemon, the trash is emptied whenever the CLI runs.
		purge := purgeitems.Process(cobra.Context(), postgrey.GetDb())
		purge.Run()
		if purge.HasErr() || noAccount[cobra] {
			return purge.GetErr()
		}

		ctx, err := cmdaccount.WithAccount(cobra.Context())
		cobra.SetContext(ctx)
		return err
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

// needsDb is false for what only prints help or shell completion: those must work with the database down.
func needsDb(command *cobra.Command) bool {
	if !command.HasParent() {
		return false
	}
	for ; command != nil; command = command.Parent() {
		switch command.Name() {
		case "completion", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
			return false
		}
	}
	return true
}

// version comes from the build (git commit, or module version without one): nothing to keep in sync by hand.
func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			return setting.Value
		}
	}
	return info.Main.Version
}

func Execute() {
	err := cmdRoot.Execute()
	if errors.Is(err, huh.ErrUserAborted) || errors.Is(err, screen.ErrQuit) {
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
