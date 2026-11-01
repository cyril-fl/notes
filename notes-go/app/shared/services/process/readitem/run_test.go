package readitem

import (
	"context"
	"errors"
	"testing"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
)

// No DB needed: makeID fails first and the loads are guarded, so a nil db is never touched.
func TestNoValue(t *testing.T) {
	process := Process(context.Background(), nil, domain.NOTE, " ")
	process.Run()
	if !errors.Is(process.GetErr(), ErrNoValue) {
		t.Fatalf("err = %v, want %v", process.GetErr(), ErrNoValue)
	}
}

// The found case is covered by editnote's test, which reads the note back.
func TestNotFound(t *testing.T) {
	env := postgrey.NewEnvironment()
	env.ReadEnv()
	db := env.Connect()
	defer db.Close()

	for _, id := range []string{"00000000-0000-0000-0000-000000000000", "nope"} {
		process := Process(context.Background(), db, domain.FOLDER, id)
		process.Run()
		if !errors.Is(process.GetErr(), ErrNotFound) {
			t.Errorf("%s: err = %v, want %v", id, process.GetErr(), ErrNotFound)
		}
	}
}
