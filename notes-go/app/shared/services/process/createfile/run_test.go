package createfile

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/silo"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
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
	var stored string
	if err := tx.NewSelect().Table("files").Column("name").Where("id = ?", file.ID).Scan(ctx, &stored); err != nil || !strings.HasPrefix(stored, secret.Prefix) {
		t.Errorf("stored name = %q, %v, want it sealed", stored, err)
	}
	if string(store.Objects[domain.UserID(ctx)+"/"+file.ID]) != "\xff\xd8\xff\xe0 jpeg" {
		t.Errorf("objects = %v", store.Objects)
	}

	unnamed := Process(ctx, tx, store, Input{Data: []byte("GIF89a...")})
	unnamed.Run()
	if unnamed.HasErr() || unnamed.GetResult().Name != "image.gif" {
		t.Errorf("unnamed = %+v, %v", unnamed.GetResult(), unnamed.GetErr())
	}

	// An attachment: a PDF, named "file" when sent without a name.
	pdf := Process(ctx, tx, store, Input{Data: []byte("%PDF-1.7 ...")})
	pdf.Run()
	if pdf.HasErr() || pdf.GetResult().Name != "file.pdf" || pdf.GetResult().ContentType != "application/pdf" {
		t.Errorf("pdf = %+v, %v", pdf.GetResult(), pdf.GetErr())
	}

	for _, test := range []struct {
		data string
		err  error
	}{
		{"", ErrNoData},
		{"<svg xmlns='http://www.w3.org/2000/svg'><script/></svg>", ErrFileType},
		{"# just text", ErrFileType},
		{"<html><script>alert(1)</script></html>", ErrFileType},
	} {
		refused := Process(ctx, tx, store, Input{Name: "x", Data: []byte(test.data)})
		refused.Run()
		if !errors.Is(refused.GetErr(), test.err) {
			t.Errorf("%q: err = %v, want %v", test.data, refused.GetErr(), test.err)
		}
	}
	if len(store.Objects) != 3 {
		t.Errorf("objects = %d, want 3", len(store.Objects))
	}

	// Through Sealed, as silo.New gives it: Silo keeps the bytes encrypted, Get gives them back.
	sealed := Process(ctx, tx, silo.Sealed(store), Input{Data: []byte("GIF89a secret")})
	sealed.Run()
	if sealed.HasErr() {
		t.Fatalf("sealed: %v", sealed.GetErr())
	}
	key := sealed.GetResult().Key()
	if strings.Contains(string(store.Objects[key]), "GIF89a") {
		t.Errorf("object in clear: %q", store.Objects[key])
	}
	body, err := silo.Sealed(store).Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if data, _ := io.ReadAll(body); string(data) != "GIF89a secret" {
		t.Errorf("Get() = %q, want the image back", data)
	}
}
