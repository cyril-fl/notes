package listtag

import (
	"context"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

type run struct {
	plumprimitive.Process
	cache

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
}

type cache struct {
	result []domain.Tag
}

func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

func (p *run) Run() {
	p.loadTags()
}

// Getters & Setters

// GetResult returns every tag, sorted case-insensitively.
func (p *run) GetResult() []domain.Tag {
	return p.cache.result
}

// Methods

// shortcut: loads every tag at once, add Limit/Offset if tags reach the thousands.
func (p *run) loadTags() {
	p.SetErr(p.db.NewSelect().
		Model(&p.cache.result).
		Where("user_id = ?", domain.UserID(p.ctx)).
		Scan(p.ctx))
	domain.SortFold(p.cache.result, func(t domain.Tag) string { return string(t.Value) }, nil)
}
