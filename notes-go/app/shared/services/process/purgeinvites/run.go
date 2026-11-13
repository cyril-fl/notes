package purgeinvites

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

// Process deletes the invitations used or expired: nothing reads them any more, and their rows would pile up.
func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

func (p *run) Run() {
	result, err := p.db.NewDelete().Model((*domain.Invite)(nil)).
		Where("used_at IS NOT NULL OR expires_at < current_timestamp").Exec(p.ctx)
	if err == nil {
		p.cache.purged, err = result.RowsAffected()
	}
	p.SetErr(err)
}

// Getters & Setters

// GetResult returns how many invitations were deleted.
func (p *run) GetResult() int64 {
	return p.cache.purged
}
