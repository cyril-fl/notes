package synclinks

import (
	"context"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/services/process/linknotes"
)

type run struct {
	plumprimitive.Process
	cache

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
}

type cache struct {
	notes   []linknotes.Input
	changed int
}

// Process resolves the [[links]] of every note, trashed ones included, and links it to the notes its text
// points to: for notes written before links were read from the text.
func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

func (p *run) Run() {
	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// No title: every note gets its own links here, relinking the notes naming it would count them twice.
		err := tx.NewRaw(`SELECT mn.item_id AS id, mn.content FROM metadata_notes AS mn`).
			Scan(ctx, &p.cache.notes)
		if err != nil {
			return err
		}

		for _, note := range p.cache.notes {
			link := linknotes.Process(ctx, tx, note)
			link.Run()
			if link.HasErr() {
				return link.GetErr()
			}
			if _, changed := link.GetResult(); changed {
				p.cache.changed++
			}
		}
		return nil
	}))
}

// Getters & Setters

// GetResult returns how many notes had their links changed: 0 on a second run.
func (p *run) GetResult() int {
	return p.cache.changed
}
