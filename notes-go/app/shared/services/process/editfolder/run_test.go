package editfolder

import (
	"context"
	"errors"
	"testing"

	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createfolder"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
)

// No DB needed: makeFolder fails first and updateFolder is guarded, so a nil db is never touched.
func TestNoValue(t *testing.T) {
	tests := map[Input]error{
		{ID: " ", Title: "T"}:  ErrNoValue,
		{ID: "id", Title: " "}: ErrNoTitle,
	}

	for input, want := range tests {
		process := Process(context.Background(), nil, input)
		process.Run()
		if !errors.Is(process.GetErr(), want) {
			t.Errorf("%+v: err = %v, want %v", input, process.GetErr(), want)
		}
	}
}

func TestEditFolder(t *testing.T) {
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

	seed := createfolder.Process(ctx, tx, createfolder.Input{Title: "Old"})
	seed.Run()
	if seed.HasErr() {
		t.Fatalf("seed: %v", seed.GetErr())
	}

	edit := Process(ctx, tx, Input{ID: seed.GetResult().ID, Title: " New "})
	edit.Run()
	if item := edit.GetResult(); edit.HasErr() || item.Title != "New" || item.UpdatedAt.IsZero() {
		t.Errorf("result = %+v, %v; want New with updated_at", item, edit.GetErr())
	}

	note := createnote.Process(ctx, tx, createnote.Input{Content: "C"})
	note.Run()
	if note.HasErr() {
		t.Fatalf("note: %v", note.GetErr())
	}

	// Last: a note id, an unknown id and a non-uuid id are all not found.
	for _, id := range []string{note.GetResult().ID, "00000000-0000-0000-0000-000000000000", "nope"} {
		missing := Process(ctx, tx, Input{ID: id, Title: "T"})
		missing.Run()
		if !errors.Is(missing.GetErr(), ErrNotFound) {
			t.Errorf("%s: err = %v, want %v", id, missing.GetErr(), ErrNotFound)
		}
	}
}
