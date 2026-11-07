package syncentities

import (
	"context"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/entities"
	"main.go/app/shared/services/notetitle"
)

// note is a note, trashed ones included.
type note struct {
	ID      string `bun:"id"`
	Title   string `bun:"title"`
	Content string `bun:"content"`
}

type run struct {
	plumprimitive.Process
	cache

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
}

type cache struct {
	notes   []note
	changed int
}

// Process repairs the notes saved by the web editor before issue #1 was fixed: & < > written as HTML
// entities in their text, and so in their title.
func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

func (p *run) Run() {
	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if err := p.loadNotes(ctx, tx); err != nil {
			return err
		}
		for _, current := range p.cache.notes {
			if err := p.syncNote(ctx, tx, current); err != nil {
				return err
			}
		}
		return nil
	}))
}

// Getters & Setters

// GetResult returns how many notes changed: 0 on a second run.
func (p *run) GetResult() int {
	return p.cache.changed
}

// Methods

func (p *run) loadNotes(ctx context.Context, tx bun.Tx) error {
	return tx.NewRaw(`
		SELECT i.id, coalesce(i.title, '') AS title, mn.content
		FROM items AS i
		JOIN metadata_notes AS mn ON mn.item_id = i.id`).
		Scan(ctx, &p.cache.notes)
}

// syncNote leaves updated_at alone: nothing the user wrote changed.
func (p *run) syncNote(ctx context.Context, tx bun.Tx, current note) error {
	content := entities.Decode(current.Content)
	if content == current.Content {
		return nil
	}
	p.cache.changed++

	_, err := tx.NewUpdate().
		Model((*domain.MetadataNote)(nil)).
		Set("content = ?", content).
		Where("item_id = ?", current.ID).
		Exec(ctx)
	if err != nil {
		return err
	}
	_, err = tx.NewUpdate().
		Model((*domain.Item)(nil)).
		Set("title = NULLIF(?, '')", notetitle.From(content)).
		Where("id = ?", current.ID).
		WhereAllWithDeleted(). // Trashed notes too: restored, they come back repaired.
		Exec(ctx)
	return err
}
