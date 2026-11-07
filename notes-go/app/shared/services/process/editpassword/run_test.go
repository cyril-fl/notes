package editpassword

import (
	"context"
	"errors"
	"testing"
	"time"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/password"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/adduser"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/testaccount"
)

func TestEditPassword(t *testing.T) {
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
	ctx, tx = testaccount.New(t, ctx, tx)

	if _, err := tx.NewUpdate().Model((*domain.User)(nil)).Set("password_hash = ?", password.Hash("old-password")).Where("id = ?", domain.UserID(ctx)).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	// This device and another one.
	this, other := password.Token(), password.Token()
	for _, token := range []string{this, other} {
		session := domain.Session{UserID: domain.UserID(ctx), TokenHash: password.TokenHash(token), ExpiresAt: time.Now().Add(time.Hour)}
		if _, err := tx.NewInsert().Model(&session).Exec(ctx); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		input Input
		want  error
	}{
		{Input{Current: "old-password", New: "short"}, adduser.ErrShortPassword},
		{Input{Current: "wrong-password", New: "new-password"}, ErrWrongPassword},
	}
	for _, test := range tests {
		process := Process(ctx, tx, test.input, this)
		process.Run()
		if !errors.Is(process.GetErr(), test.want) {
			t.Errorf("%+v: err = %v, want %v", test.input, process.GetErr(), test.want)
		}
	}

	process := Process(ctx, tx, Input{Current: "old-password", New: "new-password"}, this)
	process.Run()
	if process.HasErr() {
		t.Fatalf("edit: %v", process.GetErr())
	}

	var user domain.User
	if err := tx.NewSelect().Model(&user).Where("id = ?", domain.UserID(ctx)).Scan(ctx); err != nil || !password.Verify("new-password", user.PasswordHash) {
		t.Errorf("password not changed (%v)", err)
	}
	var left []domain.Session
	if err := tx.NewSelect().Model(&left).Where("user_id = ?", domain.UserID(ctx)).Scan(ctx); err != nil || len(left) != 1 || left[0].TokenHash != password.TokenHash(this) {
		t.Errorf("sessions = %+v (%v), want this device's only", left, err)
	}
}
