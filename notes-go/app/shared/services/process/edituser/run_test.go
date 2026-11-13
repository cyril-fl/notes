package edituser

import (
	"context"
	"errors"
	"testing"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/adduser"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createfile"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/silo"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
)

// No DB needed: checkInput fails first and updateUser is guarded, so a nil db is never touched.
func TestInvalidInput(t *testing.T) {
	text := func(s string) *string { return &s }
	days := func(n int) *int { return &n }
	tests := []struct {
		input Input
		want  error
	}{
		{Input{Locale: text("")}, ErrInvalidLocale},
		{Input{Locale: text("de")}, ErrInvalidLocale},
		{Input{Locale: text("FR")}, ErrInvalidLocale},
		{Input{Theme: text("blue")}, ErrInvalidTheme},
		{Input{Theme: text("dark")}, ErrInvalidTheme},
		{Input{ColorMode: text("dim")}, ErrInvalidColorMode},
		{Input{View: text("table")}, ErrInvalidView},
		{Input{Plan: text("max")}, ErrInvalidPlan},
		{Input{TrashRetentionDays: days(-1)}, ErrInvalidRetention},
		{Input{TrashRetentionDays: days(MaxRetentionDays + 1)}, ErrInvalidRetention},
		{Input{Name: text(" ")}, adduser.ErrNoName},
		{Input{Snippets: &[]domain.Snippet{{Name: " ", Content: "x"}}}, ErrInvalidSnippet},
		{Input{Snippets: &[]domain.Snippet{{Name: "big", Content: string(make([]byte, MaxSnippetSize+1))}}}, ErrInvalidSnippet},
		{Input{Snippets: new(make([]domain.Snippet, MaxSnippets+1))}, ErrInvalidSnippet},
		{Input{SavedSearches: &[]domain.SavedSearch{{ID: "a", Name: " "}}}, ErrInvalidSavedSearch},
		{Input{SavedSearches: &[]domain.SavedSearch{{Name: "no id"}}}, ErrInvalidSavedSearch},
		{Input{SavedSearches: &[]domain.SavedSearch{{ID: "a", Name: "x", TagMode: "both"}}}, ErrInvalidSavedSearch},
		{Input{SavedSearches: &[]domain.SavedSearch{{ID: "a", Name: "x", FolderMode: "maybe"}}}, ErrInvalidSavedSearch},
	}
	for _, test := range tests {
		process := Process(context.Background(), nil, test.input)
		process.Run()

		if !errors.Is(process.GetErr(), test.want) {
			t.Errorf("%+v: err = %v, want %v", test.input, process.GetErr(), test.want)
		}
	}
}

func TestEditUser(t *testing.T) {
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

	locale, view, never, plan := "en", "grid", 0, "pro"
	process := Process(ctx, tx, Input{Locale: &locale})
	process.Run()
	if process.HasErr() {
		t.Fatalf("edit: %v", process.GetErr())
	}
	if got := process.GetResult(); got.Locale != "en" || got.ID == "" || got.Theme != "notes" || got.ColorMode != "system" || got.TrashRetentionDays != 30 || got.Plan != "base" {
		t.Errorf("result = %+v, want locale en with ID and the default settings", got)
	}

	process = Process(ctx, tx, Input{View: &view, TrashRetentionDays: &never, Plan: &plan})
	process.Run()
	if process.HasErr() {
		t.Fatalf("edit: %v", process.GetErr())
	}
	if got := process.GetResult(); got.Locale != "en" || got.View != "grid" || got.TrashRetentionDays != 0 || got.Plan != "pro" {
		t.Errorf("result = %+v, want locale kept, view grid, retention 0, plan pro", got)
	}

	if got := process.GetResult(); len(got.Snippets) != 0 {
		t.Errorf("snippets = %+v, want none by default", got.Snippets)
	}
	snippets := []domain.Snippet{{Name: "Meeting", Content: "## Meeting #work"}, {Name: "Day", Content: "# Today"}}
	process = Process(ctx, tx, Input{Snippets: &snippets})
	process.Run()
	if got := process.GetResult(); process.HasErr() || len(got.Snippets) != 2 || got.Snippets[1] != snippets[1] {
		t.Errorf("snippets: result = %+v, err = %v; want both saved", got.Snippets, process.GetErr())
	}

	searches := []domain.SavedSearch{{ID: "s1", Name: "Work", TagMode: "all"}, {ID: "s2", Name: "Recent"}}
	process = Process(ctx, tx, Input{SavedSearches: &searches})
	process.Run()
	if got := process.GetResult(); process.HasErr() || len(got.SavedSearches) != 2 || got.SavedSearches[0].Name != "Work" || got.SavedSearches[0].TagMode != "all" {
		t.Errorf("saved searches: result = %+v, err = %v; want both saved", got.SavedSearches, process.GetErr())
	}

	process = Process(ctx, tx, Input{})
	process.Run()
	if got := process.GetResult(); process.HasErr() || got.View != "grid" || len(got.Snippets) != 2 {
		t.Errorf("empty edit: result = %+v, err = %v; want the account unchanged", got, process.GetErr())
	}

	// A column the model does not know, like one left by an old version, is not read back.
	if _, err := tx.ExecContext(ctx, "ALTER TABLE users ADD COLUMN legacy boolean NOT NULL DEFAULT false"); err != nil {
		t.Fatal(err)
	}
	process = Process(ctx, tx, Input{View: &view})
	process.Run()
	if process.HasErr() {
		t.Errorf("edit with an unknown column: %v", process.GetErr())
	}

	// The name of another account is taken, whatever its case; the savepoint keeps tx usable.
	var other domain.User
	if err := tx.NewSelect().Model(&other).Where("id <> ?", domain.UserID(ctx)).Limit(1).Scan(ctx); err != nil {
		other = domain.User{Name: "edituser-other", PasswordHash: "-"}
		if _, err := tx.NewInsert().Model(&other).Exec(ctx); err != nil {
			t.Fatal(err)
		}
	}
	taken := " " + other.Name + " "
	process = Process(ctx, tx, Input{Name: &taken})
	process.Run()
	if !errors.Is(process.GetErr(), adduser.ErrNameTaken) {
		t.Errorf("taken name: err = %v, want ErrNameTaken", process.GetErr())
	}
	name := " renamed-" + other.Name + " "
	process = Process(ctx, tx, Input{Name: &name})
	process.Run()
	if got := process.GetResult(); process.HasErr() || got.Name != "renamed-"+other.Name {
		t.Errorf("rename: result = %+v, err = %v; want the name trimmed", got, process.GetErr())
	}

	// An avatar is an image of the account: not a stranger's id, not a non-uuid, not a PDF.
	for _, bad := range []string{"not-a-uuid", "00000000-0000-0000-0000-000000000000"} {
		process = Process(ctx, tx, Input{Avatar: &bad})
		process.Run()
		if !errors.Is(process.GetErr(), ErrInvalidAvatar) {
			t.Errorf("avatar %q: err = %v, want ErrInvalidAvatar", bad, process.GetErr())
		}
	}
	file := createfile.Process(ctx, tx, silo.NewMemory(), createfile.Input{Name: "me.png", Data: []byte("\x89PNG\r\n\x1a\n")})
	file.Run()
	if file.HasErr() {
		t.Fatal(file.GetErr())
	}
	pdf := createfile.Process(ctx, tx, silo.NewMemory(), createfile.Input{Name: "me.pdf", Data: []byte("%PDF-1.7")})
	pdf.Run()
	if pdf.HasErr() {
		t.Fatal(pdf.GetErr())
	}
	pdfID := pdf.GetResult().ID
	process = Process(ctx, tx, Input{Avatar: &pdfID})
	process.Run()
	if !errors.Is(process.GetErr(), ErrInvalidAvatar) {
		t.Errorf("pdf avatar: err = %v, want ErrInvalidAvatar", process.GetErr())
	}
	avatar, none := file.GetResult().ID, ""
	process = Process(ctx, tx, Input{Avatar: &avatar})
	process.Run()
	if got := process.GetResult(); process.HasErr() || got.Avatar != avatar {
		t.Errorf("avatar: result = %+v, err = %v; want %s", got, process.GetErr(), avatar)
	}
	process = Process(ctx, tx, Input{Avatar: &none})
	process.Run()
	if got := process.GetResult(); process.HasErr() || got.Avatar != "" {
		t.Errorf("no avatar: result = %+v, err = %v; want none", got, process.GetErr())
	}
}
