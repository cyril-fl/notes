package purgeitems

import (
	"context"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/process/editnote"
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

// Process empties the trash of every account, of the items deleted longer ago than its trash_retention_days;
// a retention of 0 keeps them forever.
func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
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

// purge hard deletes: the ON DELETE CASCADE FKs remove metadata, tag links and children. The tags of the
// trashed notes are read first, then dropped if the purge left them without any note.
func (p *run) purge() {
	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		tags, err := trashedTags(ctx, tx)
		if err != nil {
			return err
		}

		result, err := tx.NewDelete().
			Model((*domain.Item)(nil)).
			ForceDelete().
			TableExpr("users AS u").
			Where("u.id = i.user_id AND u.trash_retention_days > 0").
			Where("i.deleted_at < current_timestamp - make_interval(days => u.trash_retention_days::int)").
			Exec(ctx)
		if err != nil {
			return err
		}
		if p.cache.purged, err = result.RowsAffected(); err != nil {
			return err
		}
		return editnote.DropUnusedTags(ctx, tx, tags)
	}))
}

// Helpers

// trashedTags returns the tags linked to a trashed note, of every account: the purge runs for all of them.
func trashedTags(ctx context.Context, tx bun.Tx) ([]string, error) {
	var tags []string
	err := tx.NewRaw("SELECT DISTINCT nt.tag_id FROM notes_on_tags AS nt JOIN items AS i ON i.id = nt.note_id WHERE i.deleted_at IS NOT NULL").Scan(ctx, &tags)
	return tags, err
}
