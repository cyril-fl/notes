package rotatesecret

import (
	"context"
	"testing"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/testaccount"
)

// Rewrap itself is tested in services/secret: here, keys already sealed by NOTES_SECRET stay, and one no secret
// opens stops everything.
func TestRotateSecret(t *testing.T) {
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

	const old = "b2xkIHNlY3JldCBvZiB0aGUgc2VydmVyLCAzMiBieXQ="
	process := Process(ctx, tx, old)
	process.Run()
	if process.HasErr() || process.GetResult() != 0 {
		t.Fatalf("rewrapped = %d, err = %v; want none, the key is sealed by NOTES_SECRET", process.GetResult(), process.GetErr())
	}

	if _, err := tx.NewUpdate().Model((*domain.User)(nil)).Set("data_key = 'bm90IGEga2V5'").Where("id = ?", domain.UserID(ctx)).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	broken := Process(ctx, tx, old)
	broken.Run()
	if !broken.HasErr() {
		t.Error("err = nil, want a key no secret opens to stop the rotation")
	}
}
