package seedperf

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readgraph"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
)

func TestSeed(t *testing.T) {
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

	name := fmt.Sprintf("perf-test-%d", time.Now().UnixNano())
	process := Process(ctx, tx, Input{Name: name, Password: "perf-test-password", Notes: 60})
	process.Run()
	if process.HasErr() {
		t.Fatalf("seed: %v", process.GetErr())
	}
	if created, notes := process.GetResult(); !created || notes != 60 {
		t.Fatalf("result = %v, %d, want created with 60 notes", created, notes)
	}

	var user domain.User
	if err := tx.NewSelect().Model(&user).Where("name = ?", name).Scan(ctx); err != nil {
		t.Fatal(err)
	}
	if err := secret.Load(ctx, tx, user.ID); err != nil {
		t.Fatal(err)
	}
	ctx = domain.WithUser(ctx, user.ID)

	// Read back as the app reads it: every note live, 1 to 4 tags, its title its first heading.
	graph := readgraph.Process(ctx, tx, true)
	graph.Run()
	if graph.HasErr() {
		t.Fatalf("graph: %v", graph.GetErr())
	}
	got := graph.GetResult()
	if len(got.Notes) != 60 {
		t.Fatalf("graph has %d notes, want 60", len(got.Notes))
	}
	for _, note := range got.Notes {
		if n := len(note.Tags); n < 1 || n > 4 {
			t.Errorf("%s has %d tags, want 1 to 4", note.Path, n)
		}
	}
	read := readitem.Process(ctx, tx, domain.NOTE, got.Notes[0].ID)
	read.Run()
	if read.HasErr() {
		t.Fatalf("read: %v", read.GetErr())
	}
	if note := read.GetResult(); !strings.HasPrefix(note.Content, "# "+string(note.Item.Title)+"\n") {
		t.Errorf("content %.40q does not start with its title %q", note.Content, note.Item.Title)
	}

	// An account that exists is left alone.
	again := Process(context.Background(), tx, Input{Name: name, Password: "other", Notes: 10})
	again.Run()
	if created, _ := again.GetResult(); again.HasErr() || created {
		t.Errorf("second run: created %v, err %v, want nothing done", created, again.GetErr())
	}
}
