package synctitles

import (
	"context"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/notetitle"
)

// note is a note, trashed ones included, with the title it was given by hand ("" = untitled).
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

// Process brings notes made before titles came from the text in line: a title missing from the text
// is written at its top as a heading, so no title is lost; an untitled note takes its first words.
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
	content := current.Content
	if current.Title != "" && notetitle.From(content) != current.Title {
		content = "# " + current.Title + "\n\n" + content
	}
	title := notetitle.From(content)
	if content == current.Content && title == current.Title {
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
		Set("title = NULLIF(?, '')", title).
		Where("id = ?", current.ID).
		WhereAllWithDeleted(). // Trashed notes too: restored, they must not lose their title.
		Exec(ctx)
	return err
}
