package listfiles

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
	result []domain.File
}

// Process lists the files the account keeps in Silo, the largest first: those no note cites any more too, until purgefiles,
// but not those of a purged note or of erased data (-infinity), already gone for the account.
func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

func (p *run) Run() {
	p.loadFiles()
}

// Getters & Setters

func (p *run) GetResult() []domain.File {
	return p.cache.result
}

// Methods

func (p *run) loadFiles() {
	// Not orphaned_at: time.Time cannot hold -infinity.
	p.SetErr(p.db.NewSelect().Model(&p.cache.result).ExcludeColumn("orphaned_at").
		Where("user_id = ? AND orphaned_at IS DISTINCT FROM '-infinity'", domain.UserID(p.ctx)).
		OrderExpr("size DESC, created_at DESC").Scan(p.ctx))
}
