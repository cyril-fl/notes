package synclinks

import (
	"context"
	"fmt"
	"testing"
	"time"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
)

func TestSyncLinks(t *testing.T) {
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

	u := fmt.Sprintf("test-%d", time.Now().UnixNano())
	var notes [2]string
	for i, content := range []string{"# " + u + "-target", "# " + u + "-source\n\n[[" + u + "-target]]"} {
		note := createnote.Process(ctx, tx, createnote.Input{Content: content})
		note.Run()
		if note.HasErr() {
			t.Fatalf("note %d: %v", i, note.GetErr())
		}
		notes[i] = note.GetResult().ID
	}

	// A note as it was before: its [[link]] in the text, no row.
	if _, err := tx.NewDelete().Model((*domain.NoteLink)(nil)).Where("source_id = ?", notes[1]).Exec(ctx); err != nil {
		t.Fatal(err)
	}

	first := Process(ctx, tx)
	first.Run()
	if first.HasErr() || first.GetResult() == 0 {
		t.Fatalf("first run changed %d, %v; want the source at least", first.GetResult(), first.GetErr())
	}
	linked, err := tx.NewSelect().Model((*domain.NoteLink)(nil)).Where("source_id = ? AND target_id = ?", notes[1], notes[0]).Count(ctx)
	if err != nil || linked != 1 {
		t.Errorf("link = %d, %v; want 1", linked, err)
	}

	second := Process(ctx, tx)
	second.Run()
	if second.HasErr() || second.GetResult() != 0 {
		t.Errorf("second run changed %d, %v; want 0", second.GetResult(), second.GetErr())
	}
}
