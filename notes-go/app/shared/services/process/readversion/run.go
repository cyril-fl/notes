package readversion

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

type run struct {
	plumprimitive.Process
	cache

	ctx     context.Context
	db      bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	id      string
	version string
}

type cache struct {
	version domain.NoteVersion
}

// Process reads one version of a note with its text, used by GET /api/notes/{id}/versions/{version} (#99):
// the list leaves the texts out.
func Process(ctx context.Context, db bun.IDB, id, version string) *run {
	return &run{
		ctx:     ctx,
		db:      db,
		id:      id,
		version: version,
	}
}

func (p *run) Run() {
	p.loadNote()
	p.loadVersion()
}

// Getters & Setters

func (p *run) GetResult() domain.NoteVersion {
	return p.cache.version
}

// Methods

// loadNote reuses readitem: same errors (ErrNoValue, ErrNotFound) for a missing, trashed or foreign note.
func (p *run) loadNote() {
	process := readitem.Process(p.ctx, p.db, domain.NOTE, p.id)
	process.Run()
	p.SetErr(process.GetErr())
}

// loadVersion needs no user_id check: loadNote has found the note of this account, the version must be one of its.
func (p *run) loadVersion() {
	if p.HasErr() {
		return
	}

	err := p.db.NewSelect().
		Model(&p.cache.version).
		Where("id = ? AND item_id = ?", p.version, p.id).
		Scan(p.ctx)
	if errors.Is(err, sql.ErrNoRows) {
		err = fmt.Errorf("%w: %q", readitem.ErrNotFound, p.version)
	}
	p.SetErr(err)
}
