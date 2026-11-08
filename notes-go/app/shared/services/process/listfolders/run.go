package listfolders

import (
	"context"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

// Folder is a live folder with its path from the root.
type Folder struct {
	ID   string `json:"id"`
	Path string `json:"path"`
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

// loadFolders keeps the folders domain.Paths gives, which only holds live items: a folder inside a trashed one drops out too.
// TODO: loads every folder at once, page it if folders reach the thousands.
func (p *run) loadFolders() {
	paths, err := domain.Paths(p.ctx, p.db)
	if err != nil {
		p.SetErr(err)
		return
	}
	var ids []string
	if err := p.db.NewRaw("SELECT id FROM items WHERE user_id = ? AND kind = ?", domain.UserID(p.ctx), domain.FOLDER).Scan(p.ctx, &ids); err != nil {
		p.SetErr(err)
		return
	}

	for _, id := range ids {
		if path, ok := paths[id]; ok {
			p.cache.result = append(p.cache.result, Folder{ID: id, Path: path})
		}
	}
	domain.SortFold(p.cache.result, func(f Folder) string { return f.Path }, nil)
}
