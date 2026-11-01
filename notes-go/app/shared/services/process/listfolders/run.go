package listfolders

import (
	"context"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/process/readitem"
)

// Folder is a live folder with its path from the root.
type Folder struct {
	ID   string
	Path string
}

type run struct {
	plumprimitive.Process
	cache

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
}

type cache struct {
	result []Folder
}

func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

func (p *run) Run() {
	p.loadFolders()
}

// Getters & Setters

// GetResult returns every live folder, by path.
func (p *run) GetResult() []Folder {
	return p.cache.result
}

// Methods

// loadFolders joins on paths, which only holds live items: a folder inside a trashed one drops out too.
// shortcut: loads every folder at once, page it if folders reach the thousands.
func (p *run) loadFolders() {
	p.SetErr(p.db.NewRaw(readitem.PathsCTE+`
		SELECT i.id, paths.path
		FROM items AS i JOIN paths ON paths.id = i.id
		WHERE i.kind = ?
		ORDER BY lower(paths.path)`, domain.FOLDER).
		Scan(p.ctx, &p.cache.result))
}
