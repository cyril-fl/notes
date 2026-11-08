package createsession_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/adduser"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createsession"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/deletesession"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readsession"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/resetpassword"
)

// An account from its creation to a new password: log in, find the session, log out.
func TestAccountLifecycle(t *testing.T) {
	ctx := context.Background()
	env := postgrey.NewEnvironment()
	env.ReadEnv()
	db := env.Connect()
	defer db.Close()

	schema := createschema.Process(ctx, db)
	schema.Run()
	if schema.HasErr() {
		t.Fatalf("create schema: %v", schema.GetErr())
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	// run runs a process and gives its error.
	run := func(process interface {
		Run()
		GetErr() error
	}) error {
		process.Run()
		return process.GetErr()
	}

	name := fmt.Sprintf("Account-%d", time.Now().UnixNano())
	add := adduser.Process(ctx, tx, adduser.Input{Name: name, Password: "correct horse"})
	add.Run()
	if add.HasErr() {
		t.Fatalf("add: %v", add.GetErr())
	}
	for input, want := range map[adduser.Input]error{
		{Name: name, Password: "correct horse"}: adduser.ErrNameTaken,
		{Name: " ", Password: "correct horse"}:  adduser.ErrNoName,
		{Name: name + "x", Password: "short"}:   adduser.ErrShortPassword,
	} {
		if err := run(adduser.Process(ctx, tx, input)); !errors.Is(err, want) {
			t.Errorf("add %+v: err = %v, want %v", input, err, want)
		}
	}

	for _, wrong := range []createsession.Input{{Name: name, Password: "wrong horse"}, {Name: "nobody-" + name, Password: "correct horse"}} {
		if err := run(createsession.Process(ctx, tx, wrong)); !errors.Is(err, createsession.ErrBadCredentials) {
			t.Errorf("login %q: err = %v, want %v", wrong.Name, err, createsession.ErrBadCredentials)
		}
	}
	// The name is found whatever its case.
	login := createsession.Process(ctx, tx, createsession.Input{Name: " account-" + name[8:], Password: "correct horse"})
	login.Run()
	if login.HasErr() || login.GetResult().Token == "" {
		t.Fatalf("login: %v", login.GetErr())
	}
	token := login.GetResult().Token

	read := readsession.Process(ctx, tx, token)
	read.Run()
	if read.HasErr() || read.GetResult().ID != add.GetResult().ID {
		t.Fatalf("session = %+v, %v; want the account", read.GetResult(), read.GetErr())
	}

	// A new password logs every device out.
	reset := resetpassword.Process(ctx, tx, name, "battery staple")
	reset.Run()
	if reset.HasErr() {
		t.Fatalf("reset: %v", reset.GetErr())
	}
	if err := run(readsession.Process(ctx, tx, token)); !errors.Is(err, readsession.ErrNoSession) {
		t.Errorf("session after reset: err = %v, want %v", err, readsession.ErrNoSession)
	}
	again := createsession.Process(ctx, tx, createsession.Input{Name: name, Password: "battery staple"})
	again.Run()
	if again.HasErr() {
		t.Fatalf("login with the new password: %v", again.GetErr())
	}

	if err := run(deletesession.Process(ctx, tx, again.GetResult().Token)); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if err := run(readsession.Process(ctx, tx, again.GetResult().Token)); !errors.Is(err, readsession.ErrNoSession) {
		t.Errorf("session after logout: err = %v, want %v", err, readsession.ErrNoSession)
	}
}
