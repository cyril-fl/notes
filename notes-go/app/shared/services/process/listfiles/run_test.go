package listfiles

import (
	"context"
	"testing"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createfile"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/silo"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
)

func TestListFiles(t *testing.T) {
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
	store := silo.NewMemory()

	ids := map[string]string{}
	for _, input := range []createfile.Input{
		{Name: "small.gif", Data: []byte("GIF89a")},
		{Name: "big.pdf", Data: []byte("%PDF-1.7 a longer one")},
		{Name: "erased.gif", Data: []byte("GIF89a erased")},
	} {
		create := createfile.Process(ctx, tx, store, input)
		create.Run()
		if create.HasErr() {
			t.Fatalf("create %s: %v", input.Name, create.GetErr())
		}
		ids[input.Name] = create.GetResult().ID
	}
	// As purgeaccount leaves it: gone for the account, its bytes wait for purgefiles.
	if _, err := tx.NewUpdate().Table("files").Set("orphaned_at = '-infinity'").Where("id = ?", ids["erased.gif"]).Exec(ctx); err != nil {
		t.Fatal(err)
	}

	process := Process(ctx, tx)
	process.Run()
	if process.HasErr() {
		t.Fatalf("list: %v", process.GetErr())
	}
	files := process.GetResult()
	if len(files) != 2 || files[0].Name != "big.pdf" || files[0].ContentType != "application/pdf" || files[1].Name != "small.gif" || files[1].Size != 6 {
		t.Errorf("files = %+v, want big.pdf then small.gif", files)
	}
}
