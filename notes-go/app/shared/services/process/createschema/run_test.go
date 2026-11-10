package createschema

import (
	"context"
	"testing"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
)

// A base made before a column of the model gets it back, with its default, when the schema is created again.
func TestAddColumns(t *testing.T) {
	ctx := context.Background()
	env := postgrey.NewEnvironment()
	env.ReadEnv()
	db := env.Connect()
	defer db.Close()

	schema := Process(ctx, db)
	schema.Run()
	if schema.HasErr() {
		t.Fatalf("create schema: %v", schema.GetErr())
	}
	if _, err := db.ExecContext(ctx, "ALTER TABLE users DROP COLUMN view"); err != nil {
		t.Fatal(err)
	}

	schema = Process(ctx, db)
	schema.Run()
	if schema.HasErr() {
		t.Fatalf("create schema again: %v", schema.GetErr())
	}
	var def string
	if err := db.NewRaw("SELECT column_default FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'view' AND is_nullable = 'NO'").Scan(ctx, &def); err != nil {
		t.Fatalf("users.view not back: %v", err)
	}
	if def != "'list'::character varying" {
		t.Errorf("users.view default = %s, want 'list'", def)
	}
}
