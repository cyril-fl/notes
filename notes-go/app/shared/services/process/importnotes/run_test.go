package importnotes

import (
	"context"
	"errors"
	"fmt"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createfolder"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/exportnotes"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/silo"
)

func TestImportNotes(t *testing.T) {
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

	// Unique per run: the dev database may hold anything at the root.
	create := createfolder.Process(ctx, tx, createfolder.Input{Title: fmt.Sprintf("import-%d", time.Now().UnixNano())})
	create.Run()
	if create.HasErr() {
		t.Fatal(create.GetErr())
	}
	parent := create.GetResult().ID

	dir := fstest.MapFS{
		"Projets/Idées.md":           {Data: []byte("# Idées\n\nvoir [[Plan unique 7f3]] #go")},
		"Projets/Plan.md":            {Data: []byte("# Plan unique 7f3\n\ntexte")},
		"Projets/image.png":          {Data: []byte("png")},
		"Projets/vide.md":            {Data: []byte("  \n")},
		".obsidian/config.md":        {Data: []byte("# Config")},
		"__MACOSX/Projets/._Plan.md": {Data: []byte("# Resource fork")},
		"Vide/.gitkeep":              {Data: []byte("")},
		"Racine.md":                  {Data: []byte("du texte")},
		"du lait.md":                 {Data: []byte("du lait")},
	}
	run := Process(ctx, tx, silo.NewMemory(), Input{Dir: dir, ParentID: parent})
	run.Run()
	if run.HasErr() {
		t.Fatalf("import: %v", run.GetErr())
	}
	if got, want := run.GetResult(), (Result{Folders: 2, Notes: 4}); got != want {
		t.Errorf("result = %+v, want %+v", got, want)
	}

	// The export of what came in gives the files back, the link resolved on the way and written back as typed.
	export := exportnotes.Process(ctx, tx, parent)
	export.Run()
	if export.HasErr() {
		t.Fatal(export.GetErr())
	}
	var got []string
	for _, file := range export.GetResult() {
		got = append(got, file.Path+"="+file.Content)
	}
	want := []string{
		"Projets=",
		"Projets/Idées.md=# Idées\n\nvoir [[Plan unique 7f3]] #go",
		"Projets/Plan unique 7f3.md=# Plan unique 7f3\n\ntexte",
		"Vide=",
		"du lait.md=du lait",
		"Racine.md=# Racine\n\ndu texte",
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("export = %q, want %q", got, want)
	}

	plan, _ := secret.Blind(domain.UserID(ctx), "Plan unique 7f3")
	linked, err := tx.NewSelect().Model((*domain.NoteLink)(nil)).
		Join("JOIN items AS t ON t.id = nl.target_id").Where("t.title_hash = ? AND t.parent_id IS NOT NULL", plan).
		Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if linked != 1 {
		t.Errorf("links to the imported Plan = %d, want 1", linked)
	}

	// A folder name taken stops it all: nothing of the second import stays.
	again := Process(ctx, tx, silo.NewMemory(), Input{Dir: dir, ParentID: parent})
	again.Run()
	if !errors.Is(again.GetErr(), domain.ErrFolderExists) {
		t.Errorf("second import: err = %v, want %v", again.GetErr(), domain.ErrFolderExists)
	}
}

// png is the start of a PNG file, all the type detection reads.
const png = "\x89PNG\r\n\x1a\n...."

func TestImportExportImages(t *testing.T) {
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

	create := createfolder.Process(ctx, tx, createfolder.Input{Title: fmt.Sprintf("images-%d", time.Now().UnixNano())})
	create.Run()
	if create.HasErr() {
		t.Fatal(create.GetErr())
	}
	parent := create.GetResult().ID

	// Cited four times, one image; the missing one, the web one and the text file keep their links. A ![[name]]
	// takes the file of its note's directory, or the one its path ends with, not Autre's text of the same name.
	dir := fstest.MapFS{
		"A.md":                 {Data: []byte("# A\n\n![mer](Voyage/photo%20mer.png) ![[Voyage/photo mer.png|300]] ![[absent.png]]")},
		"Voyage/Plage.md":      {Data: []byte("# Plage\n\n![sable](<photo mer.png>) ![absent](nope.png) ![web](https://x.org/y.png) ![txt](notes.txt) ![[photo mer.png]]")},
		"Autre/photo mer.png":  {Data: []byte("text")},
		"Voyage/photo mer.png": {Data: []byte(png)},
		"Voyage/notes.txt":     {Data: []byte("text")},
		"_files/old.png":       {Data: []byte(png)},
	}
	store := silo.NewMemory()
	run := Process(ctx, tx, store, Input{Dir: dir, ParentID: parent})
	run.Run()
	if run.HasErr() {
		t.Fatalf("import: %v", run.GetErr())
	}
	if got, want := run.GetResult(), (Result{Folders: 2, Notes: 2}); got != want {
		t.Errorf("result = %+v, want %+v (no folder for _files)", got, want)
	}
	if len(store.Objects) != 1 {
		t.Fatalf("objects = %d, want 1", len(store.Objects))
	}
	var file domain.File
	if err := tx.NewSelect().Model(&file).Where("user_id = ?", domain.UserID(ctx)).Scan(ctx); err != nil {
		t.Fatal(err)
	}
	if file.Name != "photo mer.png" || file.ContentType != "image/png" || store.Objects[file.Key()] == nil {
		t.Errorf("file = %+v", file)
	}

	export := exportnotes.Process(ctx, tx, parent)
	export.Run()
	if export.HasErr() {
		t.Fatal(export.GetErr())
	}
	var got []string
	for _, f := range export.GetResult() {
		got = append(got, f.Path+"="+f.Content+"="+f.Key)
	}
	image := "_files/" + file.ID + ".png"
	want := []string{
		"Autre==",
		"Voyage==",
		"Voyage/Plage.md=# Plage\n\n![sable](../" + image + ") ![absent](nope.png) ![web](https://x.org/y.png) ![txt](notes.txt) ![](../" + image + ")=",
		"A.md=# A\n\n![mer](" + image + ") ![](" + image + " \"small\") ![[absent.png]]=",
		"_files==",
		image + "==" + file.Key(),
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("export = %q, want %q", got, want)
	}

	// The second import fails on Autre, after A.md made its image: its bytes go away with the rows.
	again := Process(ctx, tx, store, Input{Dir: dir, ParentID: parent})
	again.Run()
	if !errors.Is(again.GetErr(), domain.ErrFolderExists) {
		t.Errorf("second import: err = %v, want %v", again.GetErr(), domain.ErrFolderExists)
	}
	if len(store.Objects) != 1 {
		t.Errorf("objects after a failed import = %d, want 1", len(store.Objects))
	}
}

func TestSizeTitle(t *testing.T) {
	for width, want := range map[string]string{"": "", "300": ` "small"`, "1": ` "small"`, "301": ` "medium"`, "600": ` "medium"`, "601": "", "x": ""} {
		if got := sizeTitle(width); got != want {
			t.Errorf("sizeTitle(%q) = %q, want %q", width, got, want)
		}
	}
}

func TestImportExportAttachments(t *testing.T) {
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

	create := createfolder.Process(ctx, tx, createfolder.Input{Title: fmt.Sprintf("attachments-%d", time.Now().UnixNano())})
	create.Run()
	if create.HasErr() {
		t.Fatal(create.GetErr())
	}
	parent := create.GetResult().ID

	// A link to a Word document and a ![[sound]] become attachments; a link to a note, a web link and an ![image]
	// that is a PDF keep theirs.
	dir := fstest.MapFS{
		"A.md":          {Data: []byte("# A\n\n[rapport](rapport.docx) ![[son.mp3]] [B](B.md) [web](https://x.org/a.pdf) ![pdf](doc.pdf)")},
		"B.md":          {Data: []byte("# B")},
		"rapport.docx":  {Data: []byte("PK\x03\x04 word")},
		"Audio/son.mp3": {Data: []byte("ID3 sound")},
		"doc.pdf":       {Data: []byte("%PDF-1.7")},
	}
	store := silo.NewMemory()
	run := Process(ctx, tx, store, Input{Dir: dir, ParentID: parent})
	run.Run()
	if run.HasErr() {
		t.Fatalf("import: %v", run.GetErr())
	}
	if len(store.Objects) != 2 {
		t.Fatalf("objects = %d, want 2", len(store.Objects))
	}
	files := map[string]domain.File{}
	var rows []domain.File
	if err := tx.NewSelect().Model(&rows).Where("user_id = ?", domain.UserID(ctx)).Scan(ctx); err != nil {
		t.Fatal(err)
	}
	for _, file := range rows {
		files[string(file.Name)] = file
	}
	docx, mp3 := files["rapport.docx"], files["son.mp3"]
	if docx.ContentType != "application/zip" || mp3.ContentType != "audio/mpeg" {
		t.Fatalf("files = %+v", rows)
	}

	export := exportnotes.Process(ctx, tx, parent)
	export.Run()
	if export.HasErr() {
		t.Fatal(export.GetErr())
	}
	var got []string
	for _, f := range export.GetResult() {
		got = append(got, f.Path+"="+f.Content)
	}
	docxPath, mp3Path := "_files/"+docx.ID+".docx", "_files/"+mp3.ID+".mp3"
	want := "A.md=# A\n\n[rapport](" + docxPath + ") [son.mp3](" + mp3Path + ") [B](B.md) [web](https://x.org/a.pdf) ![pdf](doc.pdf)"
	if !slices.Contains(got, want) || !slices.Contains(got, docxPath+"=") || !slices.Contains(got, mp3Path+"=") {
		t.Errorf("export = %q, want %q and both files", got, want)
	}
}
