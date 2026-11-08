package listpinned

import (
	"context"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

// Pinned is a pinned note or a favorite folder.
type Pinned struct {
	ID    string      `json:"id"`
	Kind  domain.Kind `json:"kind"`
	Title secret.Text `json:"title" swaggertype:"string"`
	Path  string      `json:"path"`
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

// Run joins live to keep live items only: a pinned note in the trash is not listed.
func (p *run) Run() {
	p.cache.result = []Pinned{}
	p.SetErr(p.db.NewRaw(domain.LiveCTE+`
		SELECT i.id, i.kind, i.title
		FROM items AS i
		JOIN live ON live.id = i.id
		LEFT JOIN metadata_notes AS mn ON mn.item_id = i.id
		LEFT JOIN metadata_folders AS mf ON mf.item_id = i.id
		WHERE mn.pinned OR mf.pinned
		ORDER BY i.id`, domain.UserID(p.ctx)).
		Scan(p.ctx, &p.cache.result))
	if p.HasErr() {
		return
	}

	paths, err := domain.Paths(p.ctx, p.db)
	p.SetErr(err)
	for i := range p.cache.result {
		p.cache.result[i].Path = paths[p.cache.result[i].ID]
	}
	domain.SortFold(p.cache.result, func(f Pinned) string { return f.Path }, nil)
}

// Getters & Setters

// GetResult returns the pinned notes and favorite folders by path.
func (p *run) GetResult() []Pinned {
	return p.cache.result
}
