package purgesessions

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
	purged int64
}

// Process deletes the expired sessions of every account: a device that never logs out (cookie cleared, device lost)
// would keep its row, and its token hash, forever.
func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

func (p *run) Run() {
	result, err := p.db.NewDelete().Model((*domain.Session)(nil)).Where("expires_at < current_timestamp").Exec(p.ctx)
	if err == nil {
		p.cache.purged, err = result.RowsAffected()
	}
	p.SetErr(err)
}

// Getters & Setters

// GetResult returns how many sessions were deleted.
func (p *run) GetResult() int64 {
	return p.cache.purged
}
