package purgeitems

import (
	"context"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
)

type run struct {
	plumprimitive.Process
	cache

	ctx           context.Context
	db            bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	retentionDays int
}

type cache struct {
	purged int64
}

// Process empties the trash of the items deleted more than retentionDays ago; retentionDays <= 0 keeps them forever.
func Process(ctx context.Context, db bun.IDB, retentionDays int) *run {
	return &run{
		ctx:           ctx,
		db:            db,
		retentionDays: retentionDays,
	}
}

func (p *run) Run() {
	p.purge()
}

// Getters & Setters

// GetResult returns how many items were purged, cascaded children not included.
func (p *run) GetResult() int64 {
	return p.cache.purged
}

// Methods

// purge hard deletes: the ON DELETE CASCADE FKs remove metadata, tag links and children.
func (p *run) purge() {
	if p.retentionDays <= 0 {
		return
	}

	result, err := p.db.NewDelete().
		Model((*domain.Item)(nil)).
		ForceDelete().
		Where("deleted_at < current_timestamp - make_interval(days => ?)", p.retentionDays).
		Exec(p.ctx)
	if err != nil {
		p.SetErr(err)
		return
	}

	p.cache.purged, err = result.RowsAffected()
	p.SetErr(err)
}
