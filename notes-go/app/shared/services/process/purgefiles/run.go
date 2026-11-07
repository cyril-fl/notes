package purgefiles

import (
	"context"
	"errors"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/silo"
)

type run struct {
	plumprimitive.Process
	cache

	ctx   context.Context
	db    bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	store silo.Store
}

type cache struct {
	orphans []domain.File
	purged  int
}

// Process deletes, for every account, the files no note cites any more, trashed notes included: a note purged
// from the trash, or an image removed from its text. An hour old at least: a pasted image waits for its note's save.
// shortcut: an image removed then brought back by undo more than an hour later stays broken, keep files longer if it bites.
func Process(ctx context.Context, db bun.IDB, store silo.Store) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		store: store,
	}
}

func (p *run) Run() {
	p.findOrphans()
	p.purge()
}

// Getters & Setters

// GetResult returns how many files were deleted.
func (p *run) GetResult() int {
	return p.cache.purged
}

// Methods

func (p *run) findOrphans() {
	err := p.db.NewSelect().Model(&p.cache.orphans).
		Where("f.created_at < current_timestamp - interval '1 hour'").
		Where(`NOT EXISTS (
			SELECT 1 FROM metadata_notes AS mn JOIN items AS i ON i.id = mn.item_id
			WHERE i.user_id = f.user_id AND strpos(mn.content, f.id::text) > 0)`).
		Scan(p.ctx)
	p.SetErr(err)
}

// purge removes the bytes before the row: a failure leaves the row, and the next purge tries again.
func (p *run) purge() {
	if p.HasErr() {
		return
	}

	var errs []error
	for _, file := range p.cache.orphans {
		err := p.store.Remove(p.ctx, file.Key())
		if err == nil {
			_, err = p.db.NewDelete().Model((*domain.File)(nil)).Where("id = ?", file.ID).Exec(p.ctx)
		}
		if err != nil {
			errs = append(errs, err)
			continue
		}
		p.cache.purged++
	}
	p.SetErr(errors.Join(errs...))
}
