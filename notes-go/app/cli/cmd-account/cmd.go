// Package cmdaccount logs the CLI in and out, and lets the admin create accounts and reset their password.
package cmdaccount

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/huhform"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/adduser"
	"main.go/app/shared/services/process/createsession"
	"main.go/app/shared/services/process/deletesession"
	"main.go/app/shared/services/process/readsession"
	"main.go/app/shared/services/process/resetpassword"
)

// sessionFile keeps the CLI's session token, readable by the user only: it opens the account.
func sessionFile() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "notes", "session"), nil
}

func readToken() string {
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

// Login asks the account's name and password and keeps the session in the user's config folder.
func Login(cobra *cobra.Command, args []string) error {
	var input createsession.Input
	if err := huhform.New(
		huh.NewInput().Title("Name").Value(&input.Name),
		huh.NewInput().Title("Password").EchoMode(huh.EchoModePassword).Value(&input.Password),
	).Run(); err != nil {
		return err
	}

	process := createsession.Process(cobra.Context(), postgrey.GetDb(), input)
	process.Run()
	if process.HasErr() {
		return process.GetErr()
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

// Logout ends the CLI's session, in the base and on disk.
func Logout(cobra *cobra.Command, args []string) error {
	process := deletesession.Process(cobra.Context(), postgrey.GetDb(), readToken())
	process.Run()
	if process.HasErr() {
		return process.GetErr()
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
