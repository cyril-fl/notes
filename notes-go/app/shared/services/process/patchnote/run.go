package patchnote

import (
	"context"
	"errors"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/hashtag"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/editnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/pinnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

var ErrLocked = errors.New("note is read-only, unlock it first")

// Input has pointers to tell a field left out from an empty one: editnote replaces everything,
// so a change without tags must keep them, not clear them.
type Input struct {
	ID       string
	Pinned   *bool
	Readonly *bool
	Content  *string
	Tags     *[]string
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
// content are the tags of the note, unless the tags are given: they then rewrite it. Pinning or locking alone
// leaves the note as it is, updated_at included. A locked note refuses content and tags, unless unlocked in the same call.
// TODO: the read is not locked, a concurrent edit can still be lost, fine for one user; SELECT … FOR UPDATE if that changes.
func Process(ctx context.Context, db bun.IDB, input Input) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		input: input,
	}
}

// Run reads, pins, locks then edits in one transaction, so a failed edit leaves the pin and lock as they were (#39).
func (p *run) Run() {
	// p.ctx, not the one RunInTx hands: a bun.Tx's own ctx may act for another account.
	p.SetErr(p.db.RunInTx(p.ctx, nil, func(_ context.Context, tx bun.Tx) error {
		for _, step := range []func(bun.Tx) error{p.readNote, p.pinNote, p.lockNote, p.editNote} {
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

func (p *run) lockNote(tx bun.Tx) error {
	if p.input.Readonly == nil {
		return nil
	}

	// readNote has found the note of this account: no user_id check needed here.
	_, err := tx.NewUpdate().
		Model((*domain.MetadataNote)(nil)).
		Set("readonly = ?", *p.input.Readonly).
		Where("item_id = ?", p.cache.note.Item.ID).
		Exec(p.ctx)
	if err != nil {
		return err
	}
	p.cache.note.Readonly = *p.input.Readonly
	return nil
}

func (p *run) editNote(tx bun.Tx) error {
	if p.input.Content == nil && p.input.Tags == nil {
		return nil
	}
	if p.cache.note.Readonly {
		return ErrLocked
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
