package emptytrash

import (
	"context"
	"errors"
	"fmt"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
)

var ErrNotFound = errors.New("item not found in the trash")

type run struct {
	plumprimitive.Process

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	ids []string
}

// Process deletes for good, whatever its age, what the trash holds: everything, or the removals of ids only.
func Process(ctx context.Context, db bun.IDB, ids ...string) *run {
	return &run{
		ctx: ctx,
		db:  db,
		ids: ids,
	}
}

func (p *run) Run() {
	p.empty()
}

// Methods

// empty hard deletes: the ON DELETE CASCADE FKs remove metadata, tag links and children. Only trashed items:
// a live id is not found. id::text, so an id that is not a uuid is not found rather than a SQL error.
func (p *run) empty() {
	query := p.db.NewDelete().
		Model((*domain.Item)(nil)).
		ForceDelete().
		Where("deleted_at IS NOT NULL AND user_id = ?", domain.UserID(p.ctx))
	if len(p.ids) > 0 {
		query = query.Where("id::text IN (?)", bun.List(p.ids))
	}

	result, err := query.Exec(p.ctx)
	if err != nil {
		p.SetErr(err)
		return
	}
	if rows, _ := result.RowsAffected(); len(p.ids) > 0 && rows < int64(len(p.ids)) {
		p.SetErr(fmt.Errorf("%w: %q", ErrNotFound, p.ids))
	}
}
