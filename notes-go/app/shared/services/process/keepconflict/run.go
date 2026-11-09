package keepconflict

import (
	"context"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

type run struct {
	plumprimitive.Process
	cache

	ctx     context.Context
	db      bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	id      string
	content string
}

type cache struct {
	note    readitem.Result
	version domain.NoteVersion
}

// Process keeps a text an edit could not save, the note having changed on another device (409), as a version marked
// conflict, used by POST /api/notes/{id}/versions (#78). The same text as the note's keeps nothing.
func Process(ctx context.Context, db bun.IDB, id, content string) *run {
	return &run{
		ctx:     ctx,
		db:      db,
		id:      id,
		content: content,
	}
}

func (p *run) Run() {
	p.loadNote()
	p.keepVersion()
}

// Getters & Setters

// GetResult returns the version kept, empty when the text was the note's.
func (p *run) GetResult() domain.NoteVersion {
	return p.cache.version
}

// Methods

// loadNote reuses readitem: same errors (ErrNoValue, ErrNotFound) for a missing, trashed or foreign note.
func (p *run) loadNote() {
	process := readitem.Process(p.ctx, p.db, domain.NOTE, p.id)
	process.Run()
	if process.HasErr() {
		p.SetErr(process.GetErr())
		return
	}
	p.cache.note = process.GetResult()
}

func (p *run) keepVersion() {
	if p.HasErr() || p.content == p.cache.note.Content {
		return
	}

	version, err := domain.NewNoteVersion(p.ctx, p.cache.note.Item.ID, p.content)
	if err != nil {
		p.SetErr(err)
		return
	}
	version.Conflict = true
	p.SetErr(p.db.NewInsert().Model(&version).Returning("id, created_at").Scan(p.ctx))
	version.Content = secret.Text(p.content) // answered as listversions does, in clear
	p.cache.version = version
}
