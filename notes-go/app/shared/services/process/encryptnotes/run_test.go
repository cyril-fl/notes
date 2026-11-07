package encryptnotes

import (
	"context"
	"strings"
	"testing"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createfile"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/secret"
	"main.go/app/shared/services/silo"
	"main.go/app/shared/services/testaccount"
)

func TestEncryptNotes(t *testing.T) {
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

	file := createfile.Process(ctx, tx, silo.NewMemory(), createfile.Input{Name: "a", Data: []byte("\x89PNG\r\n\x1a\n")})
	file.Run()
	note := createnote.Process(ctx, tx, createnote.Input{Content: "# Old\n\n![](" + domain.FileURL(file.GetResult().ID) + ")"})
	note.Run()
	if file.HasErr() || note.HasErr() {
		t.Fatal(file.GetErr(), note.GetErr())
	}
	id, user := note.GetResult().ID, domain.UserID(ctx)

	// What createnote saves is sealed already.
	var raw string
	if err := tx.NewRaw("SELECT content FROM metadata_notes WHERE item_id = ?", id).Scan(ctx, &raw); err != nil || !strings.HasPrefix(raw, secret.Prefix) {
		t.Fatalf("content = %q, %v; want it sealed", raw, err)
	}

	// As written before encryption: the text in clear, no notes_on_files, the snippets a JSON array.
	_, err = tx.NewRaw("UPDATE metadata_notes SET content = ? WHERE item_id = ?", "# Old\n\n![]("+domain.FileURL(file.GetResult().ID)+")", id).Exec(ctx)
	if err == nil {
		_, err = tx.NewRaw("DELETE FROM notes_on_files WHERE note_id = ?", id).Exec(ctx)
	}
	if err == nil {
		_, err = tx.NewRaw(`UPDATE users SET snippets = '[{"name":"Day","content":"# Today"}]' WHERE id = ?`, user).Exec(ctx)
	}
	if err != nil {
		t.Fatal(err)
	}

	process := Process(ctx, tx)
	process.Run()
	if process.HasErr() || process.GetResult().Notes == 0 || process.GetResult().Accounts == 0 {
		t.Fatalf("result = %+v, err = %v; want the note and the account sealed", process.GetResult(), process.GetErr())
	}

	var snippets string
	err = tx.NewRaw("SELECT mn.content, u.snippets::text FROM metadata_notes AS mn, users AS u WHERE mn.item_id = ? AND u.id = ?", id, user).Scan(ctx, &raw, &snippets)
	if err != nil || !strings.HasPrefix(raw, secret.Prefix) || strings.Contains(snippets, "Today") {
		t.Errorf("content = %q, snippets = %q, err = %v; want both sealed", raw, snippets, err)
	}
	var account domain.User
	if err := tx.NewSelect().Model(&account).Where("id = ?", user).Scan(ctx); err != nil || len(account.Snippets) != 1 || account.Snippets[0].Content != "# Today" {
		t.Errorf("snippets = %+v, err = %v; want them opened", account.Snippets, err)
	}
	if n, err := tx.NewSelect().Model((*domain.NoteFile)(nil)).Where("note_id = ?", id).Count(ctx); err != nil || n != 1 {
		t.Errorf("notes_on_files = %d, %v; want the image linked", n, err)
	}

	again := Process(ctx, tx)
	again.Run()
	if again.HasErr() || again.GetResult() != (Result{}) {
		t.Errorf("second run = %+v, err = %v; want nothing left", again.GetResult(), again.GetErr())
	}
}
