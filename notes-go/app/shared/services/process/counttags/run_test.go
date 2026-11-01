package counttags

import (
	"context"
	"fmt"
	"testing"
	"time"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/createtag"
	"main.go/app/shared/services/process/deleteitems"
)

func TestCountTags(t *testing.T) {
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

	used, unused := fmt.Sprintf("test-%d-used", time.Now().UnixNano()), fmt.Sprintf("test-%d-unused", time.Now().UnixNano())
	seed := createtag.Process(ctx, tx, []string{unused})
	seed.Run()
	var trashedID string
	for _, title := range []string{"A", "B", "Trashed"} {
		note := createnote.Process(ctx, tx, createnote.Input{Title: title, Content: "C", Tags: []string{used}})
		note.Run()
		if note.HasErr() {
			t.Fatalf("note %s: %v", title, note.GetErr())
		}
		trashedID = note.GetResult().ID
	}
	trash := deleteitems.Process(ctx, tx, domain.NOTE, []string{trashedID})
	trash.Run()
	if seed.HasErr() || trash.HasErr() {
		t.Fatalf("seed: %v %v", seed.GetErr(), trash.GetErr())
	}

	process := Process(ctx, tx)
	process.Run()
	counts := map[string]int{}
	for _, tag := range process.GetResult() {
		counts[tag.Value] = tag.Notes
	}
	if process.HasErr() || counts[used] != 2 || counts[unused] != 0 {
		t.Errorf("counts = used %d, unused %d, %v; want 2 and 0", counts[used], counts[unused], process.GetErr())
	}
	if _, ok := counts[unused]; !ok {
		t.Error("unused tag missing")
	}
}
