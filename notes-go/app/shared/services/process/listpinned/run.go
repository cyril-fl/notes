package listpinned

import (
	"context"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
)

// Pinned is a pinned note.
type Pinned struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

type run struct {
	plumprimitive.Process
	cache

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
}

type cache struct {
	result []Pinned
}

func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

// Run joins paths to keep live notes only: a pinned note in the trash is not listed.
func (p *run) Run() {
	p.cache.result = []Pinned{}
	p.SetErr(p.db.NewRaw(domain.PathsCTE+`
		SELECT i.id, coalesce(i.title, '') AS title, paths.path
		FROM items AS i
		JOIN paths ON paths.id = i.id
		JOIN metadata_notes AS mn ON mn.item_id = i.id
		WHERE i.kind = ? AND mn.pinned
		ORDER BY lower(paths.path), i.id`, domain.NOTE).
		Scan(p.ctx, &p.cache.result))
}

// Getters & Setters

// GetResult returns the pinned notes by path.
func (p *run) GetResult() []Pinned {
	return p.cache.result
}
