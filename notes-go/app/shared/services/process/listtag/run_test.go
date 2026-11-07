package listtag

import (
	"context"
	"fmt"
	"main.go/app/shared/services/testaccount"
	"testing"
	"time"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/createtag"
)

func TestListSortedCaseInsensitive(t *testing.T) {
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
	prefix := fmt.Sprintf("test-%d-", time.Now().UnixNano())
	b, a := prefix+"b", prefix+"A"

	create := createtag.Process(ctx, tx, []string{b, a})
	create.Run()
	if create.HasErr() {
		t.Fatalf("create: %v", create.GetErr())
	}

	process := Process(ctx, tx)
	process.Run()
	if process.HasErr() {
		t.Fatalf("list: %v", process.GetErr())
	}

	// Other tags may exist in the DB: only check ours appear, A before b.
	if got := indexOf(process.GetResult(), a); got == -1 || got > indexOf(process.GetResult(), b) {
		t.Errorf("result = %+v, want %s before %s", process.GetResult(), a, b)
	}
}

func indexOf(tags []domain.Tag, value string) int {
	for i, tag := range tags {
		if tag.Value == value {
			return i
		}
	}
	return -1
}
