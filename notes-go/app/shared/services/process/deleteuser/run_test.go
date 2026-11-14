package deleteuser

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/adduser"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/silo"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
	"github.com/uptrace/bun"
)

func TestDeleteUser(t *testing.T) {
	ctx, tx := base(t)
	store := silo.NewMemory()

	// The account to delete, with an image of its own.
	victim := add(t, ctx, tx, "victim")
	// No name: it is sealed by the key of its account, which this test never loads.
	file := domain.File{UserID: victim.ID, ContentType: "image/png", Size: 3}
	if _, err := tx.NewInsert().Model(&file).Returning("id").Exec(ctx); err != nil {
		t.Fatal(err)
	}
	if err := store.Put(ctx, file.Key(), []byte("png"), file.ContentType); err != nil {
		t.Fatal(err)
	}
	// Sealed snippets and saved searches, then its key forgotten as after a restart: the delete must not read them.
	if err := secret.Load(ctx, tx, victim.ID); err != nil {
		t.Fatal(err)
	}
	snippets, err := domain.Snippets{{Name: "sig", Content: "victim"}}.Seal(victim.ID)
	if err != nil {
		t.Fatal(err)
	}
	searches, err := domain.SavedSearches{{ID: "s1", Name: "Go"}}.Seal(victim.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.NewUpdate().Model((*domain.User)(nil)).Set("snippets = ?", snippets).Set("saved_searches = ?", searches).Where("id = ?", victim.ID).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	secret.Forget(victim.ID)

	process := Process(ctx, tx, store, victim.ID)
	process.Run()
	if process.HasErr() {
		t.Fatalf("delete: %v", process.GetErr())
	}

	gone, err := tx.NewSelect().Model((*domain.User)(nil)).Where("id = ?", victim.ID).Exists(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if gone {
		t.Error("the account is still there")
	}
	if _, left := store.Objects[file.Key()]; left {
		t.Error("the bytes of its image are still in the store")
	}
}

func TestDeleteUserRefusals(t *testing.T) {
	ctx, tx := base(t)
	store := silo.NewMemory()

	// Itself: the admin would lock the base out, and its own danger zone already empties its data.
	process := Process(ctx, tx, store, domain.UserID(ctx))
	process.Run()
	if !errors.Is(process.GetErr(), ErrSelf) {
		t.Errorf("its own account: err = %v, want ErrSelf", process.GetErr())
	}

	process = Process(ctx, tx, store, "3f2a4b1c-0d9e-4a8b-9c7d-1e2f3a4b5c6d")
	process.Run()
	if !errors.Is(process.GetErr(), ErrNotFound) {
		t.Errorf("an unknown id: err = %v, want ErrNotFound", process.GetErr())
	}

	// The last admin: nobody could create an account any more.
	admin := add(t, ctx, tx, "admin")
	if _, err := tx.NewUpdate().Model((*domain.User)(nil)).Set("admin = (id = ?)", admin.ID).Where("TRUE").Exec(ctx); err != nil {
		t.Fatal(err)
	}
	process = Process(ctx, tx, store, admin.ID)
	process.Run()
	if !errors.Is(process.GetErr(), ErrLastAdmin) {
		t.Errorf("the last admin: err = %v, want ErrLastAdmin", process.GetErr())
	}
}

// connected is the base of the package: one connection and one schema for every test, the schema taking locks
// that a second run would wait for.
var connected = sync.OnceValue(func() *bun.DB {
	env := postgrey.NewEnvironment()
	env.ReadEnv()
	db := env.Connect()
	schema := createschema.Process(context.Background(), db)
	schema.Run()
	if schema.HasErr() {
		panic(schema.GetErr())
	}
	return db
})

// base gives a transaction rolled back at the end, with an account in ctx.
func base(t *testing.T) (context.Context, bun.Tx) {
	t.Helper()
	ctx := context.Background()
	db := connected()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { tx.Rollback() })
	ctx, tx = testaccount.New(t, ctx, tx)
	return ctx, tx
}

func add(t *testing.T, ctx context.Context, tx bun.Tx, name string) domain.User {
	t.Helper()
	process := adduser.Process(ctx, tx, adduser.Input{Name: name + "-" + t.Name(), Password: "a-good-password"})
	process.Run()
	if process.HasErr() {
		t.Fatalf("add %s: %v", name, process.GetErr())
	}
	return process.GetResult()
}
