package listrecent

import (
	"context"
	"time"

	"main.go/app/shared/services/secret"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
)

// limit is how many notes the list shows.
const limit = 20

// Recent is a note and when it last changed.
type Recent struct {
	ID        string      `json:"id"`
	Title     secret.Text `json:"title" swaggertype:"string"`
	Path      string      `json:"path"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type run struct {
	plumprimitive.Process
	cache

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
}

type cache struct {
	result []Recent
}

func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

func (p *run) Run() {
	p.loadNotes()
}

// Getters & Setters

// GetResult returns the last changed notes, most recent first.
func (p *run) GetResult() []Recent {
	return p.cache.result
}

// Methods

// loadNotes falls back on created_at: a note never edited has no updated_at. Joining live keeps live notes only.
func (p *run) loadNotes() {
	p.SetErr(p.db.NewRaw(domain.LiveCTE+`
		SELECT i.id, i.title, coalesce(i.updated_at, i.created_at) AS updated_at
		FROM items AS i
		JOIN live ON live.id = i.id
		WHERE i.kind = ?
		ORDER BY updated_at DESC, i.id
		LIMIT ?`,
		domain.UserID(p.ctx), domain.NOTE, limit).
		Scan(p.ctx, &p.cache.result))
	if p.HasErr() {
		return
	}

	paths, err := domain.Paths(p.ctx, p.db)
	p.SetErr(err)
	for i := range p.cache.result {
		p.cache.result[i].Path = paths[p.cache.result[i].ID]
	}
	domain.SortFold(p.cache.result, func(r Recent) string { return r.Path }, func(a, b Recent) int { return b.UpdatedAt.Compare(a.UpdatedAt) })
}
