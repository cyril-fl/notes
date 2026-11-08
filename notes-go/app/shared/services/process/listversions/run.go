package listversions

import (
	"context"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

type run struct {
	plumprimitive.Process
	cache

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	id  string
}

type cache struct {
	versions []domain.NoteVersion
}

// Process lists the texts a note had before its edits, newest first, used by GET /api/notes/{id}/versions (#61).
func Process(ctx context.Context, db bun.IDB, id string) *run {
	return &run{
		ctx: ctx,
		db:  db,
		id:  id,
	}
}

func (p *run) Run() {
	p.loadNote()
	p.loadVersions()
}

// Getters & Setters

func (p *run) GetResult() []domain.NoteVersion {
	return p.cache.versions
}

// Methods

// loadNote reuses readitem: same errors (ErrNoValue, ErrNotFound) for a missing, trashed or foreign note.
func (p *run) loadNote() {
	process := readitem.Process(p.ctx, p.db, domain.NOTE, p.id)
	process.Run()
	p.SetErr(process.GetErr())
}

// loadVersions needs no user_id check: loadNote has found the note of this account.
func (p *run) loadVersions() {
	if p.HasErr() {
		return
	}

	p.SetErr(p.db.NewSelect().
		Model(&p.cache.versions).
		Where("item_id = ?", p.id).
		OrderExpr("created_at DESC").
		Scan(p.ctx))
}
