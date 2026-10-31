package edittag

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/createtag"
)

// No DB needed: makeValues fails first and updateTag is guarded, so a nil db is never touched.
func TestBlankValues(t *testing.T) {
	for _, values := range [][2]string{{"", "go"}, {"go", "  "}} {
		process := Process(context.Background(), nil, values[0], values[1])
		process.Run()

		if !errors.Is(process.GetErr(), ErrNoValue) {
			t.Errorf("%q: err = %v, want %v", values, process.GetErr(), ErrNoValue)
		}
	}
}

func TestRename(t *testing.T) {
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

	// Unique per run so real data in the DB cannot collide.
	value := fmt.Sprintf("test-%d", time.Now().UnixNano())
	other := value + "-other"

	create := createtag.Process(ctx, tx, []string{value, other})
	create.Run()
	if create.HasErr() {
		t.Fatalf("create: %v", create.GetErr())
	}

	// Case-only rename, matched case-insensitively.
	upper := "TEST" + value[4:]
	rename := Process(ctx, tx, " "+value+" ", upper)
	rename.Run()
	if rename.HasErr() {
		t.Fatalf("rename: %v", rename.GetErr())
	}
	if got := rename.GetResult(); got.Value != upper || got.ID == "" {
		t.Errorf("rename result = %+v, want value %s with ID", got, upper)
	}

	unknown := Process(ctx, tx, value+"-unknown", "x")
	unknown.Run()
	if !errors.Is(unknown.GetErr(), ErrNotFound) {
		t.Errorf("unknown: err = %v, want %v", unknown.GetErr(), ErrNotFound)
	}

	// Last: the unique violation aborts the transaction.
	collision := Process(ctx, tx, other, value)
	collision.Run()
	if !errors.Is(collision.GetErr(), ErrAlreadyExists) {
		t.Errorf("collision: err = %v, want %v", collision.GetErr(), ErrAlreadyExists)
	}
}
