// Package cmdaccount logs the CLI in and out, empties the account, and lets the admin create accounts, reset their
// password and change the server's secret.
package cmdaccount

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"charm.land/huh/v2"
	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/huhform"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/adduser"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createsession"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/deletesession"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/purgeaccount"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readsession"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/resetpassword"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/rotatesecret"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/seedperf"
	"github.com/spf13/cobra"
)

// sessionFile keeps the CLI's session token, readable by the user only: it opens the account.
func sessionFile() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "notes", "session"), nil
}

// readToken takes NOTES_SESSION first, for an agent that cannot run notes login where it runs, then the session file.
func readToken() string {
	if token := strings.TrimSpace(os.Getenv("NOTES_SESSION")); token != "" {
		return token
	}
	file, err := sessionFile()
	if err != nil {
		return ""
	}
	token, _ := os.ReadFile(file)
	return strings.TrimSpace(string(token))
}

// WithAccount gives ctx the account of the CLI's session: every command but these here acts for it.
func WithAccount(ctx context.Context) (context.Context, error) {
	session := readsession.Process(ctx, postgrey.GetDb(), readToken())
	session.Run()
	if errors.Is(session.GetErr(), readsession.ErrNoSession) {
		return ctx, fmt.Errorf("%w: run notes login", domain.ErrNoUser)
	}
	if session.HasErr() {
		return ctx, session.GetErr()
	}
	return domain.WithUser(ctx, session.GetResult().ID), nil
}

// Login asks the account's name and password and keeps the session in the user's config folder, or with
// --print-token prints a session of its own, for NOTES_SESSION, and keeps nothing.
func Login(cobra *cobra.Command, args []string) error {
	var input createsession.Input
	print, _ := cobra.Flags().GetBool("print-token")
	form := huhform.New(
		huh.NewInput().Title("Name").Value(&input.Name),
		huh.NewInput().Title("Password").EchoMode(huh.EchoModePassword).Value(&input.Password),
	)
	// The form on stderr leaves stdout to the token alone, for NOTES_SESSION=$(notes login --print-token).
	if print && huhform.Input == nil {
		form = form.WithOutput(os.Stderr)
	}
	if err := form.Run(); err != nil {
		return err
	}

	process := createsession.Process(cobra.Context(), postgrey.GetDb(), input)
	process.Run()
	if process.HasErr() {
		return process.GetErr()
	}
	if print {
		fmt.Println(process.GetResult().Token)
		return nil
	}

	file, err := sessionFile()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(file, []byte(process.GetResult().Token), 0o600); err != nil {
		return err
	}
	fmt.Printf("Logged in as %s\n", process.GetResult().User.Name)
	return nil
}

// Logout ends the CLI's session, in the base and on disk; a NOTES_SESSION one in the base only, the file is another.
func Logout(cobra *cobra.Command, args []string) error {
	process := deletesession.Process(cobra.Context(), postgrey.GetDb(), readToken())
	process.Run()
	if process.HasErr() {
		return process.GetErr()
	}
	if strings.TrimSpace(os.Getenv("NOTES_SESSION")) != "" {
		fmt.Println("Logged out")
		return nil
	}
	if file, err := sessionFile(); err == nil {
		if err := os.Remove(file); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	fmt.Println("Logged out")
	return nil
}

// AddUser creates an account.
func AddUser(cobra *cobra.Command, args []string) error {
	password, err := newPassword(cobra)
	if err != nil {
		return err
	}

	process := adduser.Process(cobra.Context(), postgrey.GetDb(), adduser.Input{Name: args[0], Password: password})
	process.Run()
	if process.HasErr() {
		return process.GetErr()
	}
	fmt.Printf("Successfully created account %s\n", process.GetResult().Name)
	return nil
}

// SeedPerf creates an account of generated notes, 50,000 unless args[1] says otherwise, to measure the app at scale.
func SeedPerf(cobra *cobra.Command, args []string) error {
	notes := 50000
	if len(args) == 2 {
		var err error
		if notes, err = strconv.Atoi(args[1]); err != nil || notes < 1 {
			return fmt.Errorf("notes: %q is not a positive number", args[1])
		}
	}
	password, err := newPassword(cobra)
	if err != nil {
		return err
	}

	process := seedperf.Process(cobra.Context(), postgrey.GetDb(), seedperf.Input{Name: args[0], Password: password, Notes: notes})
	process.Run()
	if process.HasErr() {
		return process.GetErr()
	}
	if created, count := process.GetResult(); created {
		fmt.Printf("Successfully created account %s with %d notes\n", args[0], count)
	} else {
		fmt.Printf("Account %s exists: nothing done\n", args[0])
	}
	return nil
}

// ResetPassword gives an account a new password and logs all its devices out: for a forgotten one.
func ResetPassword(cobra *cobra.Command, args []string) error {
	password, err := newPassword(cobra)
	if err != nil {
		return err
	}

	process := resetpassword.Process(cobra.Context(), postgrey.GetDb(), args[0], password)
	process.Run()
	if process.HasErr() {
		return process.GetErr()
	}
	fmt.Printf("Successfully reset the password of %s: log in again on every device\n", args[0])
	return nil
}

// RotateSecret seals the keys of the accounts by NOTES_SECRET, in place of NOTES_OLD_SECRET.
func RotateSecret(cobra *cobra.Command, args []string) error {
	process := rotatesecret.Process(cobra.Context(), postgrey.GetDb(), os.Getenv("NOTES_OLD_SECRET"))
	process.Run()
	if process.HasErr() {
		return process.GetErr()
	}
	fmt.Printf("Successfully sealed %d account keys by the new NOTES_SECRET: put it in .env.docker too, then make api\n", process.GetResult())
	return nil
}

// Purge deletes for good every note, folder, tag and image of the account logged in, once its name is typed.
func Purge(cobra *cobra.Command, args []string) error {
	session := readsession.Process(cobra.Context(), postgrey.GetDb(), readToken())
	session.Run()
	if session.HasErr() {
		return session.GetErr()
	}
	name := session.GetResult().Name

	var typed string
	if err := huhform.New(
		huh.NewInput().Title(fmt.Sprintf("Type %s to delete every note, folder, tag and image for good", name)).Value(&typed).
			Validate(sameName(name)),
	).Run(); err != nil {
		return err
	}

	process := purgeaccount.Process(cobra.Context(), postgrey.GetDb())
	process.Run()
	if process.HasErr() {
		return process.GetErr()
	}
	fmt.Printf("Successfully emptied account %s\n", name)
	return nil
}

// sameName accepts the account's name only, exactly as the app's danger zone does.
func sameName(name string) func(string) error {
	return func(value string) error {
		if value != name {
			return errors.New("not the account's name")
		}
		return nil
	}
}

// newPassword reads it from stdin with --password-stdin (scripts), else asks it twice.
func newPassword(cobra *cobra.Command) (string, error) {
	if fromStdin, _ := cobra.Flags().GetBool("password-stdin"); fromStdin {
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && line == "" {
			return "", err
		}
		return strings.TrimRight(line, "\r\n"), nil
	}

	var password, again string
	err := huhform.New(
		huh.NewInput().Title("Password").EchoMode(huh.EchoModePassword).Value(&password),
		huh.NewInput().Title("Password again").EchoMode(huh.EchoModePassword).Value(&again).
			Validate(func(value string) error {
				if value != password {
					return errors.New("the passwords differ")
				}
				return nil
			}),
	).Run()
	return password, err
}
