package deletetag

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

// No DB needed: makeValues fails first and deleteTags is guarded, so a nil db is never touched.
func TestBlankValues(t *testing.T) {
	process := Process(context.Background(), nil, []string{"", "   "})
	process.Run()

	if !errors.Is(process.GetErr(), ErrNoValue) {
		t.Fatalf("err = %v, want %v", process.GetErr(), ErrNoValue)
	}
}

func TestDeleteIgnoresCaseAndUnknown(t *testing.T) {
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

	create := createtag.Process(ctx, tx, []string{value, value + "-keep"})
	create.Run()
	if create.HasErr() {
		t.Fatalf("create: %v", create.GetErr())
	}

	first := Process(ctx, tx, []string{" TEST" + value[4:] + " ", value + "-unknown"})
	first.Run()
	if first.HasErr() {
		t.Fatalf("first run: %v", first.GetErr())
	}
	got := first.GetResult()
	if len(got) != 1 || got[0].Value != value {
		t.Fatalf("first result = %+v, want [%s]", got, value)
	}

	second := Process(ctx, tx, []string{value})
	second.Run()
	if second.HasErr() {
		t.Fatalf("second run: %v", second.GetErr())
	}
	if len(second.GetResult()) != 0 {
		t.Errorf("second result = %+v, want empty", second.GetResult())
	}

	// The other tag must survive.
	keep := Process(ctx, tx, []string{value + "-keep"})
	keep.Run()
	if keep.HasErr() || len(keep.GetResult()) != 1 {
		t.Errorf("keep result = %+v, err = %v, want 1 deleted", keep.GetResult(), keep.GetErr())
	}
}
