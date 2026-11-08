package cmdaccount

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/adduser"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createsession"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readsession"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/resetpassword"
	"github.com/spf13/cobra"
)

func TestSameName(t *testing.T) {
	check := sameName("cyril")
	if err := check("cyril"); err != nil {
		t.Errorf("the name: %v", err)
	}
	for _, typed := range []string{"", "Cyril", "cyril ", "cyri"} {
		if check(typed) == nil {
			t.Errorf("%q confirmed the purge", typed)
		}
	}
}

// The commands commit on postgrey.GetDb(), not in a rolled-back transaction: each account made here is deleted
// at the end. The forms of huh need a terminal: login and the purge's confirmation are left out.
func setup(t *testing.T) context.Context {
	t.Helper()
	ctx := context.Background()
	env := postgrey.NewEnvironment()
	env.ReadEnv()
	db := env.Connect()
	schema := createschema.Process(ctx, db)
	schema.Run()
	if schema.HasErr() {
		t.Fatalf("create schema: %v", schema.GetErr())
	}
	previous := postgrey.GetDb()
	postgrey.SetDb(db)
	t.Cleanup(func() {
		postgrey.SetDb(previous)
		db.Close()
	})

	// os.UserConfigDir: HOME on macOS, XDG_CONFIG_HOME on Linux. The session file of the user stays untouched.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	return ctx
}

func account(t *testing.T, ctx context.Context) string {
	t.Helper()
	name := fmt.Sprintf("test-%d", time.Now().UnixNano())
	t.Cleanup(func() {
		if _, err := postgrey.GetDb().NewDelete().Model((*domain.User)(nil)).Where("name = ?", name).Exec(ctx); err != nil {
			t.Errorf("delete %s: %v", name, err)
		}
	})
	return name
}

// run calls a command as cobra does, the password given on stdin, and returns what it printed.
func run(t *testing.T, ctx context.Context, runE func(*cobra.Command, []string) error, stdin string, args ...string) (string, error) {
	t.Helper()
	command := &cobra.Command{}
	command.Flags().Bool("password-stdin", true, "")
	command.SetContext(ctx)

	in, inWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	inWriter.WriteString(stdin)
	inWriter.Close()
	out, outWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdinBefore, stdoutBefore := os.Stdin, os.Stdout
	os.Stdin, os.Stdout = in, outWriter
	defer func() { os.Stdin, os.Stdout = stdinBefore, stdoutBefore }()

	runErr := runE(command, args)
	outWriter.Close()
	printed, _ := io.ReadAll(out)
	return string(printed), runErr
}

func logIn(t *testing.T, ctx context.Context, name, password string) error {
	t.Helper()
	process := createsession.Process(ctx, postgrey.GetDb(), createsession.Input{Name: name, Password: password})
	process.Run()
	return process.GetErr()
}

func TestAddUser(t *testing.T) {
	ctx := setup(t)
	name := account(t, ctx)

	printed, err := run(t, ctx, AddUser, "a password\n", name)
	if err != nil || printed != "Successfully created account "+name+"\n" {
		t.Fatalf("printed %q, err = %v; want the account created", printed, err)
	}
	if err := logIn(t, ctx, name, "a password"); err != nil {
		t.Errorf("log in: %v; want the password read from stdin, without its newline", err)
	}

	for label, test := range map[string]struct {
		stdin string
		want  error
	}{
		"name taken":     {"another password\n", adduser.ErrNameTaken},
		"short password": {"short\n", adduser.ErrShortPassword},
		"no password":    {"", io.EOF},
	} {
		printed, err := run(t, ctx, AddUser, test.stdin, name)
		if !errors.Is(err, test.want) || printed != "" {
			t.Errorf("%s: printed %q, err = %v; want %v", label, printed, err, test.want)
		}
	}
}

func TestResetPassword(t *testing.T) {
	ctx := setup(t)
	name := account(t, ctx)
	if _, err := run(t, ctx, AddUser, "the old password\n", name); err != nil {
		t.Fatal(err)
	}

	printed, err := run(t, ctx, ResetPassword, "the new password\r\n", name)
	if err != nil || !strings.HasPrefix(printed, "Successfully reset the password of "+name) {
		t.Fatalf("printed %q, err = %v; want the password reset", printed, err)
	}
	if err := logIn(t, ctx, name, "the new password"); err != nil {
		t.Errorf("log in with the new password: %v", err)
	}
	if err := logIn(t, ctx, name, "the old password"); !errors.Is(err, createsession.ErrBadCredentials) {
		t.Errorf("log in with the old password: err = %v, want %v", err, createsession.ErrBadCredentials)
	}

	if _, err := run(t, ctx, ResetPassword, "the new password\n", name+"-nobody"); !errors.Is(err, resetpassword.ErrNotFound) {
		t.Errorf("unknown account: err = %v, want %v", err, resetpassword.ErrNotFound)
	}
}

// shortcut: the success only, a key no secret opens would have to be committed in the shared test base; the
// error path is tested in process/rotatesecret.
func TestRotateSecret(t *testing.T) {
	ctx := setup(t)
	t.Setenv("NOTES_OLD_SECRET", "b2xkIHNlY3JldCBvZiB0aGUgc2VydmVyLCAzMiBieXQ=")

	printed, err := run(t, ctx, RotateSecret, "")
	if err != nil || !strings.HasPrefix(printed, "Successfully sealed ") {
		t.Errorf("printed %q, err = %v; want the keys sealed", printed, err)
	}
}

// Logged out, the commands acting for an account refuse; logged in, they act for its user; logout ends both.
func TestSession(t *testing.T) {
	ctx := setup(t)

	if _, err := WithAccount(ctx); !errors.Is(err, domain.ErrNoUser) {
		t.Errorf("no session: err = %v, want %v", err, domain.ErrNoUser)
	}
	if _, err := run(t, ctx, Purge, ""); !errors.Is(err, readsession.ErrNoSession) {
		t.Errorf("purge with no session: err = %v, want %v", err, readsession.ErrNoSession)
	}

	name := account(t, ctx)
	if _, err := run(t, ctx, AddUser, "a password\n", name); err != nil {
		t.Fatal(err)
	}
	session := createsession.Process(ctx, postgrey.GetDb(), createsession.Input{Name: name, Password: "a password"})
	session.Run()
	if session.HasErr() {
		t.Fatal(session.GetErr())
	}
	file, err := sessionFile()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(session.GetResult().Token+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	withAccount, err := WithAccount(ctx)
	if err != nil || domain.UserID(withAccount) != session.GetResult().User.ID {
		t.Fatalf("user = %q, err = %v; want %q", domain.UserID(withAccount), err, session.GetResult().User.ID)
	}

	printed, err := run(t, ctx, Logout, "")
	if err != nil || printed != "Logged out\n" {
		t.Fatalf("printed %q, err = %v; want logged out", printed, err)
	}
	if _, err := os.Stat(file); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("session file: %v, want it removed", err)
	}
	check := readsession.Process(ctx, postgrey.GetDb(), session.GetResult().Token)
	check.Run()
	if !errors.Is(check.GetErr(), readsession.ErrNoSession) {
		t.Errorf("session in the base: err = %v, want it deleted", check.GetErr())
	}
}
