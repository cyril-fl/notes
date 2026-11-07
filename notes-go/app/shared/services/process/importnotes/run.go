package importnotes

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"strings"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/services/notetitle"
	"main.go/app/shared/services/process/createfolder"
	"main.go/app/shared/services/process/createnote"
)

// Input is what to import: every directory and .md file of Dir, into the folder ParentID ("" = the root).
type Input struct {
	Dir      fs.FS
	ParentID string
}

// Result counts what was created.
type Result struct {
	Folders int `json:"folders"`
	Notes   int `json:"notes"`
}

type run struct {
	plumprimitive.Process
	cache

	ctx   context.Context
	db    bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	input Input
}

type cache struct {
	folders map[string]string // directory path -> folder id
	result  Result
}

// Process imports markdown files as notes and their directories as folders, all or nothing: a folder
// name already taken stops it. Titles, #tags and [[links]] come from the text, as for any note.
func Process(ctx context.Context, db bun.IDB, input Input) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		input: input,
	}
}

func (p *run) Run() {
	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		p.cache = cache{folders: map[string]string{".": p.input.ParentID}}
		return fs.WalkDir(p.input.Dir, ".", func(name string, entry fs.DirEntry, err error) error {
			if err != nil || name == "." {
				return err
			}
			return p.importEntry(ctx, tx, name, entry)
		})
	}))
}

// Getters & Setters

func (p *run) GetResult() Result {
	return p.cache.result
}

// Methods

// importEntry skips hidden entries (.git, .obsidian…), files other than .md and blank notes.
func (p *run) importEntry(ctx context.Context, tx bun.Tx, name string, entry fs.DirEntry) error {
	hidden := strings.HasPrefix(entry.Name(), ".")
	if entry.IsDir() {
		if hidden {
			return fs.SkipDir
		}
		return p.importFolder(ctx, tx, name, entry.Name())
	}
	if hidden || !strings.EqualFold(path.Ext(name), ".md") {
		return nil
	}

	data, err := fs.ReadFile(p.input.Dir, name)
	if err != nil || strings.TrimSpace(string(data)) == "" {
		return err
	}
	note := createnote.Process(ctx, tx, createnote.Input{Content: withTitle(string(data), name), ParentID: p.cache.folders[path.Dir(name)]})
	note.Run()
	if note.HasErr() {
		return fmt.Errorf("%s: %w", name, note.GetErr())
	}
	p.cache.result.Notes++
	return nil
}

// withTitle writes the file name as a heading on a text without one (Obsidian titles a note by its file),
// unless its first words already give that name: an untitled note exported by notes comes back as it was.
func withTitle(content, name string) string {
	title := strings.TrimSuffix(path.Base(name), path.Ext(name))
	if notetitle.Heading(content) != "" || notetitle.From(content) == title {
		return content
	}
	return "# " + title + "\n\n" + content
}

func (p *run) importFolder(ctx context.Context, tx bun.Tx, name, title string) error {
	folder := createfolder.Process(ctx, tx, createfolder.Input{Title: title, ParentID: p.cache.folders[path.Dir(name)]})
	folder.Run()
	if folder.HasErr() {
		return fmt.Errorf("%s: %w", name, folder.GetErr())
	}
	p.cache.folders[name] = folder.GetResult().ID
	p.cache.result.Folders++
	return nil
}
