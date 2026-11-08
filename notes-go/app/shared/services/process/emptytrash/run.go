package emptytrash

import (
	"context"
	"errors"
	"fmt"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/editnote"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
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
// The tags of the trashed notes are read first, then dropped if the delete left them without any note.
func (p *run) empty() {
	var rows int64
	err := p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var tags []string
		err := tx.NewRaw("SELECT DISTINCT nt.tag_id FROM notes_on_tags AS nt JOIN items AS i ON i.id = nt.note_id WHERE i.deleted_at IS NOT NULL AND i.user_id = ?", domain.UserID(ctx)).Scan(ctx, &tags)
		if err != nil {
			return err
		}

		query := tx.NewDelete().
			Model((*domain.Item)(nil)).
			ForceDelete().
			Where("deleted_at IS NOT NULL AND user_id = ?", domain.UserID(ctx))
		if len(p.ids) > 0 {
			query = query.Where("id::text IN (?)", bun.List(p.ids))
		}

		result, err := query.Exec(ctx)
		if err != nil {
			return err
		}
		rows, _ = result.RowsAffected()
		return editnote.DropUnusedTags(ctx, tx, tags)
	})
	if err != nil {
		p.SetErr(err)
		return
	}
	if len(p.ids) > 0 && rows < int64(len(p.ids)) {
		p.SetErr(fmt.Errorf("%w: %q", ErrNotFound, p.ids))
	}
}
