package patchnote

import (
	"context"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/hashtag"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/editnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/pinnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

// Input has pointers to tell a field left out from an empty one: editnote replaces everything,
// so a change without tags must keep them, not clear them.
type Input struct {
	ID      string
	Pinned  *bool
	Content *string
	Tags    *[]string
}

type run struct {
	plumprimitive.Process
	cache

	ctx   context.Context
	db    bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	input Input
}

type cache struct {
	note readitem.Result
}

// Process changes the fields given of a note, used by PATCH /api/notes/{id} and `edit note`. The #tags of the
// content are the tags of the note, unless the tags are given: they then rewrite it. Pinning alone leaves
// the note as it is, updated_at included.
// shortcut: the read is not locked, a concurrent edit can still be lost, fine for one user; SELECT … FOR UPDATE if that changes.
func Process(ctx context.Context, db bun.IDB, input Input) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		input: input,
	}
}

// Run reads, pins then edits in one transaction, so a failed edit leaves the pin as it was (#39).
func (p *run) Run() {
	// p.ctx, not the one RunInTx hands: a bun.Tx's own ctx may act for another account.
	p.SetErr(p.db.RunInTx(p.ctx, nil, func(_ context.Context, tx bun.Tx) error {
		for _, step := range []func(bun.Tx) error{p.readNote, p.pinNote, p.editNote} {
			if err := step(tx); err != nil {
				return err
			}
		}
		return nil
	}))
}

// Getters & Setters

// GetResult returns the note as GET /api/notes/{id} answers it.
func (p *run) GetResult() readitem.Result {
	return p.cache.note
}

// Methods

func (p *run) readNote(tx bun.Tx) error {
	process := readitem.Process(p.ctx, tx, domain.NOTE, p.input.ID)
	process.Run()
	if process.HasErr() {
		return process.GetErr()
	}
	p.cache.note = process.GetResult()
	return nil
}

func (p *run) pinNote(tx bun.Tx) error {
	if p.input.Pinned == nil {
		return nil
	}

	process := pinnote.Process(p.ctx, tx, domain.NOTE, p.input.ID, *p.input.Pinned)
	process.Run()
	if process.HasErr() {
		return process.GetErr()
	}
	p.cache.note.Pinned = *p.input.Pinned
	return nil
}

func (p *run) editNote(tx bun.Tx) error {
	if p.input.Content == nil && p.input.Tags == nil {
		return nil
	}

	input := editnote.Input{ID: p.input.ID, Content: p.cache.note.Content}
	if p.input.Content != nil {
		input.Content = *p.input.Content
	}
	input.Tags = hashtag.Parse(input.Content)
	if p.input.Tags != nil {
		input.Tags = *p.input.Tags
	}

	process := editnote.Process(p.ctx, tx, input)
	process.Run()
	if process.HasErr() {
		return process.GetErr()
	}
	// The saved text is read back: editnote has turned its [[links]] into note ids.
	return p.readNote(tx)
}
