package cmd

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"charm.land/huh/v2"
	cmdaccount "github.com/cyril-fl/notes/notes-go/app/cli/cmd-account"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/purgeitems"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/screen"
	"github.com/cyril-fl/notes/notes-go/config"
	"github.com/spf13/cobra"
)

var c = config.Current
var cmd = c.Commands

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

// groups sorts the commands of the help by scope; a command left out shows under "Additional Commands".
var groups = []struct {
	group    cobra.Group
	commands []string
}{
	{cobra.Group{ID: "notes", Title: "Notes, folders and tags:"}, []string{"add", "show", "edit", "move", "remove"}},
	{cobra.Group{ID: "browse", Title: "Browse:"}, []string{"finder", "recent", "search", "graph"}},
	{cobra.Group{ID: "trash", Title: "Trash:"}, []string{"trash", "restore"}},
	{cobra.Group{ID: "files", Title: "Markdown files:"}, []string{"import", "export"}},
	{cobra.Group{ID: "account", Title: "Account:"}, []string{"login", "logout", "account"}},
	{cobra.Group{ID: "admin", Title: "Admin (on the server):"}, []string{"user", "reset-password", "rotate-secret", "serve"}},
}

func groupCommands() {
	byName := map[string]string{}
	for _, g := range groups {
		cmdRoot.AddGroup(&g.group)
		for _, name := range g.commands {
			byName[name] = g.group.ID
		}
	}
	for _, command := range cmdRoot.Commands() {
		command.GroupID = byName[command.Name()]
	}
}

func Execute() {
	groupCommands()
	err := cmdRoot.Execute()
	if errors.Is(err, huh.ErrUserAborted) || errors.Is(err, screen.ErrQuit) {
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
