package listtag

import (
	"context"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
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
		OrderExpr("lower(value)").
		Scan(p.ctx))
}
