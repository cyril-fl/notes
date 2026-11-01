package listfolders

import (
	"context"
	"testing"

	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/createfolder"
	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/createschema"
	"main.go/app/shared/services/process/deleteitems"
)

func TestListFolders(t *testing.T) {
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

	folder := func(title, parentID string) string {
		process := createfolder.Process(ctx, tx, createfolder.Input{Title: title, ParentID: parentID})
		process.Run()
		if process.HasErr() {
			t.Fatalf("folder %s: %v", title, process.GetErr())
		}
		return process.GetResult().ID
	}
	top := folder("ListTop", "")
	child := folder("Child", top)
	trashed := folder("Trashed", top)
	folder("Hidden", trashed)
	note := createnote.Process(ctx, tx, createnote.Input{Title: "Note", Content: "C", ParentID: top})
	note.Run()
	trash := deleteitems.Process(ctx, tx, domain.FOLDER, []string{trashed})
	trash.Run()
	if note.HasErr() || trash.HasErr() {
		t.Fatalf("seed: %v %v", note.GetErr(), trash.GetErr())
	}

	process := Process(ctx, tx)
	process.Run()
	paths := map[string]string{}
	for _, f := range process.GetResult() {
		paths[f.Path] = f.ID
	}
	if process.HasErr() || paths["/ListTop"] != top || paths["/ListTop/Child"] != child {
		t.Errorf("paths = %v, %v; want /ListTop and /ListTop/Child", paths, process.GetErr())
	}
	for _, gone := range []string{"/ListTop/Note", "/ListTop/Trashed", "/ListTop/Trashed/Hidden"} {
		if _, ok := paths[gone]; ok {
			t.Errorf("%s listed, want only live folders", gone)
		}
	}
}
