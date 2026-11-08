package createtag

import (
	"context"
	"errors"
	"fmt"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
	"testing"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
)

// No DB needed: makeTags fails first and insertTags is guarded, so a nil db is never touched.
func TestBlankValues(t *testing.T) {
	process := Process(context.Background(), nil, []string{"", "   "})
	process.Run()

	if !errors.Is(process.GetErr(), ErrNoValue) {
		t.Fatalf("err = %v, want %v", process.GetErr(), ErrNoValue)
	}
}

func TestCreateSkipsExisting(t *testing.T) {
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

	// Unique per run so real data in the DB cannot collide.
	value := fmt.Sprintf("test-%d", time.Now().UnixNano())

	first := Process(ctx, tx, []string{value, " " + value + "-b ", "TEST" + value[4:]})
	first.Run()
	if first.HasErr() {
		t.Fatalf("first run: %v", first.GetErr())
	}
	got := first.GetResult()
	if len(got) != 2 || string(got[0].Value) != value || string(got[1].Value) != value+"-b" {
		t.Fatalf("first result = %+v, want [%s %s-b]", got, value, value)
	}
	if got[0].ID == "" {
		t.Errorf("ID not returned")
	}

	second := Process(ctx, tx, []string{value})
	second.Run()
	if second.HasErr() {
		t.Fatalf("second run: %v", second.GetErr())
	}
	if len(second.GetResult()) != 0 {
		t.Errorf("second result = %+v, want empty", second.GetResult())
	}
}
