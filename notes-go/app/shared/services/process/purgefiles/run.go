package purgefiles

import (
	"context"
	"errors"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/silo"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
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

// cited is true of a file some note of its account cites, trashed notes and their versions included, or that is its avatar.
const cited = `(EXISTS (SELECT 1 FROM notes_on_files AS nf WHERE nf.file_id = f.id)
	OR EXISTS (SELECT 1 FROM note_versions AS nv WHERE f.id = ANY(nv.file_ids))
	OR EXISTS (SELECT 1 FROM users AS au WHERE au.avatar = f.id))`

// Process deletes the files no note cites any more. Those of a note deleted for good go at once (createschema's
// trigger dates them -infinity), unless another note still cites them. Those removed from a text wait the trash
// retention of their account, like a trashed item: undo or the link pasted again brings them back until then.
// A retention of 0 keeps those forever.
func Process(ctx context.Context, db bun.IDB, store silo.Store) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		store: store,
	}
}

func (p *run) Run() {
	p.markOrphans()
	p.findOrphans()
	p.purge()
}

// Getters & Setters

// GetResult returns how many files were deleted.
func (p *run) GetResult() int {
	return p.cache.purged
}

// Methods

// markOrphans starts the clock of the files no longer cited, and stops it for those cited again.
func (p *run) markOrphans() {
	_, err := p.db.NewUpdate().Model((*domain.File)(nil)).Set("orphaned_at = NULL").
		Where("orphaned_at IS NOT NULL AND " + cited).Exec(p.ctx)
	if err == nil {
		_, err = p.db.NewUpdate().Model((*domain.File)(nil)).Set("orphaned_at = current_timestamp").
			Where("orphaned_at IS NULL AND NOT " + cited).Exec(p.ctx)
	}
	p.SetErr(err)
}

func (p *run) findOrphans() {
	if p.HasErr() {
		return
	}

	// Not orphaned_at: time.Time cannot hold -infinity.
	err := p.db.NewSelect().Model(&p.cache.orphans).Column("f.id", "f.user_id").
		Join("JOIN users AS u ON u.id = f.user_id").
		Where(`f.orphaned_at = '-infinity' OR (u.trash_retention_days > 0
			AND f.orphaned_at < current_timestamp - make_interval(days => u.trash_retention_days::int))`).
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
