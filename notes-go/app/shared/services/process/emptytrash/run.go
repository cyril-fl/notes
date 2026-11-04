package emptytrash

import (
	"context"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
)

type run struct {
	plumprimitive.Process

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
}

// Process deletes everything in the trash for good, whatever its age.
func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

func (p *run) Run() {
	p.empty()
}

// Methods

// empty hard deletes: the ON DELETE CASCADE FKs remove metadata, tag links and children.
func (p *run) empty() {
	_, err := p.db.NewDelete().
		Model((*domain.Item)(nil)).
		ForceDelete().
		Where("deleted_at IS NOT NULL").
		Exec(p.ctx)
	p.SetErr(err)
}
