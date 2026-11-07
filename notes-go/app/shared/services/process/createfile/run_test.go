package createfile

import (
	"context"
	"errors"
	"testing"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/silo"
	"main.go/app/shared/services/testaccount"
)

func TestCreateFile(t *testing.T) {
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

	// The type comes from the bytes, the name loses its path.
	create := Process(ctx, tx, store, Input{Name: `C:\photos\chat.jpg`, Data: []byte("\xff\xd8\xff\xe0 jpeg")})
	create.Run()
	if create.HasErr() {
		t.Fatalf("create: %v", create.GetErr())
	}
	file := create.GetResult()
	if file.ID == "" || file.Name != "chat.jpg" || file.ContentType != "image/jpeg" || file.Size != 9 {
		t.Errorf("file = %+v", file)
	}
	if string(store.Objects[domain.UserID(ctx)+"/"+file.ID]) != "\xff\xd8\xff\xe0 jpeg" {
		t.Errorf("objects = %v", store.Objects)
	}

	unnamed := Process(ctx, tx, store, Input{Data: []byte("GIF89a...")})
	unnamed.Run()
	if unnamed.HasErr() || unnamed.GetResult().Name != "image.gif" {
		t.Errorf("unnamed = %+v, %v", unnamed.GetResult(), unnamed.GetErr())
	}

	for _, test := range []struct {
		data string
		err  error
	}{
		{"", ErrNoData},
		{"<svg xmlns='http://www.w3.org/2000/svg'><script/></svg>", ErrNotImage},
		{"# just text", ErrNotImage},
	} {
		refused := Process(ctx, tx, store, Input{Name: "x", Data: []byte(test.data)})
		refused.Run()
		if !errors.Is(refused.GetErr(), test.err) {
			t.Errorf("%q: err = %v, want %v", test.data, refused.GetErr(), test.err)
		}
	}
	if len(store.Objects) != 2 {
		t.Errorf("objects = %d, want 2", len(store.Objects))
	}
}
