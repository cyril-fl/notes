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

	"charm.land/huh/v2"
	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/huhform"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/adduser"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createsession"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createtag"
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
// at the end.
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

// keys types each field of the next huh forms, then waits: the form focuses the next field a key late, a key typed
// at once would land in the previous one.
func keys(t *testing.T, fields ...string) {
	t.Helper()
	reader, writer := io.Pipe()
	go func() {
		for _, field := range fields {
			if _, err := writer.Write([]byte(field)); err != nil {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
	huhform.Input = reader
	t.Cleanup(func() {
		huhform.Input = nil
		writer.Close()
	})
}

// run calls a command as cobra does and returns what it printed. The password comes from stdin, or from the form
// once keys are given.
func run(t *testing.T, ctx context.Context, runE func(*cobra.Command, []string) error, stdin string, args ...string) (string, error) {
	t.Helper()
	command := &cobra.Command{}
	command.Flags().Bool("password-stdin", huhform.Input == nil, "")
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

func TestAskPassword(t *testing.T) {
	ctx := setup(t)
	name := account(t, ctx)

	keys(t, "a password\r", "another one\r", "\x03")
	if _, err := run(t, ctx, AddUser, "", name); !errors.Is(err, huh.ErrUserAborted) {
		t.Fatalf("passwords that differ: err = %v, want the form kept until aborted", err)
	}
	keys(t, "a password\r", "a password\r")
	if printed, err := run(t, ctx, AddUser, "", name); err != nil || printed != "Successfully created account "+name+"\n" {
		t.Fatalf("printed %q, err = %v; want the account created", printed, err)
	}
	if err := logIn(t, ctx, name, "a password"); err != nil {
		t.Errorf("log in: %v; want the password typed twice", err)
	}
}

func TestRotateSecret(t *testing.T) {
	ctx := setup(t)
	t.Setenv("NOTES_OLD_SECRET", "b2xkIHNlY3JldCBvZiB0aGUgc2VydmVyLCAzMiBieXQ=")

	printed, err := run(t, ctx, RotateSecret, "")
	if err != nil || !strings.HasPrefix(printed, "Successfully sealed ") {
		t.Errorf("printed %q, err = %v; want the keys sealed", printed, err)
	}

	// A key no secret opens stops the rotation.
	name := account(t, ctx)
	broken := domain.User{Name: name, PasswordHash: "-", DataKey: "bm90IGEga2V5"}
	if _, err := postgrey.GetDb().NewInsert().Model(&broken).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	if printed, err := run(t, ctx, RotateSecret, ""); err == nil || printed != "" {
		t.Errorf("printed %q, err = %v; want the broken key to stop the rotation", printed, err)
	}
}

func TestLogin(t *testing.T) {
	ctx := setup(t)
	name := account(t, ctx)
	if _, err := run(t, ctx, AddUser, "a password\n", name); err != nil {
		t.Fatal(err)
	}
	file, err := sessionFile()
	if err != nil {
		t.Fatal(err)
	}

	keys(t, name+"\r", "the wrong one\r")
	if _, err := run(t, ctx, Login, ""); !errors.Is(err, createsession.ErrBadCredentials) {
		t.Errorf("wrong password: err = %v, want %v", err, createsession.ErrBadCredentials)
	}
	if _, err := os.Stat(file); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("session file after a wrong password: %v, want none", err)
	}

	keys(t, name+"\r", "a password\r")
	printed, err := run(t, ctx, Login, "")
	if err != nil || printed != "Logged in as "+name+"\n" {
		t.Fatalf("printed %q, err = %v; want logged in", printed, err)
	}
	if info, err := os.Stat(file); err != nil || info.Mode().Perm() != 0o600 {
		t.Errorf("session file: %v, %v; want it readable by the user only", info, err)
	}
	if _, err := WithAccount(ctx); err != nil {
		t.Errorf("act for the account: %v; want the session kept", err)
	}
}

// The purge asks the account's name: a wrong one empties nothing, the right one every tag (and note, folder, image).
func TestPurge(t *testing.T) {
	ctx := setup(t)
	name := account(t, ctx)
	if _, err := run(t, ctx, AddUser, "a password\n", name); err != nil {
		t.Fatal(err)
	}
	keys(t, name+"\r", "a password\r")
	if _, err := run(t, ctx, Login, ""); err != nil {
		t.Fatal(err)
	}
	ctx, err := WithAccount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	tag := createtag.Process(ctx, postgrey.GetDb(), []string{"kept"})
	tag.Run()
	if tag.HasErr() {
		t.Fatal(tag.GetErr())
	}
	tags := func() int64 {
		count, err := postgrey.GetDb().NewSelect().Model((*domain.Tag)(nil)).Where("user_id = ?", domain.UserID(ctx)).Count(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return count
	}

	keys(t, "not "+name+"\r", "\x03")
	if _, err := run(t, ctx, Purge, ""); !errors.Is(err, huh.ErrUserAborted) || tags() != 1 {
		t.Fatalf("wrong name: err = %v, %d tags; want the form kept until aborted and the tag kept", err, tags())
	}
	keys(t, name+"\r")
	printed, err := run(t, ctx, Purge, "")
	if err != nil || printed != "Successfully emptied account "+name+"\n" || tags() != 0 {
		t.Errorf("printed %q, err = %v, %d tags; want the account emptied", printed, err, tags())
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

// --print-token prints a session of its own and keeps no file; NOTES_SESSION then wins over the file, and logout ends
// that session only, the file's staying.
func TestPrintToken(t *testing.T) {
	ctx := setup(t)
	name := account(t, ctx)
	if _, err := run(t, ctx, AddUser, "a password\n", name); err != nil {
		t.Fatal(err)
	}
	keys(t, name+"\r", "a password\r")
	if _, err := run(t, ctx, Login, ""); err != nil {
		t.Fatal(err)
	}
	file, err := sessionFile()
	if err != nil {
		t.Fatal(err)
	}
	kept, _ := os.ReadFile(file)

	keys(t, name+"\r", "a password\r")
	command := func(cobra *cobra.Command, args []string) error {
		cobra.Flags().Bool("print-token", true, "")
		return Login(cobra, args)
	}
	printed, err := run(t, ctx, command, "")
	token := strings.TrimSpace(printed)
	if err != nil || token == "" || token == string(kept) {
		t.Fatalf("printed %q, err = %v; want a new token", printed, err)
	}
	if now, _ := os.ReadFile(file); string(now) != string(kept) {
		t.Errorf("session file rewritten by --print-token")
	}

	t.Setenv("NOTES_SESSION", token+"\n")
	if readToken() != token {
		t.Errorf("readToken() = %q, want NOTES_SESSION's", readToken())
	}
	if _, err := WithAccount(ctx); err != nil {
		t.Errorf("act through NOTES_SESSION: %v", err)
	}
	if _, err := run(t, ctx, Logout, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := WithAccount(ctx); !errors.Is(err, domain.ErrNoUser) {
		t.Errorf("NOTES_SESSION after logout: err = %v, want %v", err, domain.ErrNoUser)
	}

	t.Setenv("NOTES_SESSION", "")
	if _, err := WithAccount(ctx); err != nil {
		t.Errorf("the file's session after logging NOTES_SESSION out: %v, want it kept", err)
	}
}
