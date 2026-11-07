package editlocale

import (
	"context"
	"errors"
	"testing"

	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/testaccount"
)

// No DB needed: checkLocale fails first and updateUser is guarded, so a nil db is never touched.
func TestInvalidLocale(t *testing.T) {
	for _, locale := range []string{"", "de", "FR"} {
		process := Process(context.Background(), nil, locale)
		process.Run()

		if !errors.Is(process.GetErr(), ErrInvalidLocale) {
			t.Errorf("%q: err = %v, want %v", locale, process.GetErr(), ErrInvalidLocale)
		}
	}
}

func TestEditLocale(t *testing.T) {
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

	process := Process(ctx, tx, "en")
	process.Run()
	if process.HasErr() {
		t.Fatalf("edit: %v", process.GetErr())
	}
	if got := process.GetResult(); got.Locale != "en" || got.ID == "" {
		t.Errorf("result = %+v, want locale en with ID", got)
	}
}
