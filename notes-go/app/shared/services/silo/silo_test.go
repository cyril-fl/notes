package silo_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/silo"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
)

// The real Silo of compose.yaml, skipped when it is not running: a file put comes back, sealed, and goes.
func TestSilo(t *testing.T) {
	t.Setenv("S3_BUCKET", "notes-test") // like the base of the tests, not the bucket the API serves
	ctx := context.Background()
	reach, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	store, err := silo.New(reach)
	if err != nil {
		t.Skipf("Silo not running: %v", err)
	}

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
	ctx, _ = testaccount.New(t, ctx, tx)

	key := fmt.Sprintf("%s/test-%d", domain.UserID(ctx), time.Now().UnixNano())
	data := []byte("an image")
	if err := store.Put(ctx, key, data, "image/png"); err != nil {
		t.Fatalf("put: %v", err)
	}
	defer store.Remove(ctx, key)

	body, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	got, _ := io.ReadAll(body)
	body.Close()
	if !bytes.Equal(got, data) {
		t.Errorf("get = %q, want %q", got, data)
	}

	// purgefiles sweeps what List gives: the key must come back as written, and dated, not a prefix nor a folder.
	objects, err := store.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	listed := false
	for _, object := range objects {
		listed = listed || object.Key == key && time.Since(object.Written) < time.Hour
	}
	if !listed {
		t.Errorf("list = %+v, want %s written just now", objects, key)
	}

	if err := store.Remove(ctx, key); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if _, err := store.Get(ctx, key); err == nil {
		t.Error("get after remove: err = nil, want the file gone")
	}
	if err := store.Remove(ctx, key); err != nil {
		t.Errorf("remove a missing file: %v, want nil", err)
	}
}
