package patchnote

import (
	"context"
	"errors"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/hashtag"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/editnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/pinnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

var (
	ErrLocked = errors.New("note is read-only, unlock it first")
	ErrStale  = errors.New("note changed elsewhere since it was read")
)

// A text replaced is kept at most once every versionEvery per note, the versionsKept last ones (#61).
const (
	versionEvery = "5 minutes"
	versionsKept = 50
)

// Input has pointers to tell a field left out from an empty one: editnote replaces everything,
// so a change without tags must keep them, not clear them.
type Input struct {
	ID       string
	Pinned   *bool
	Readonly *bool
	Content  *string
	Tags     *[]string
	// UpdatedAt is the version the text was written over: content and tags are refused once the note changed since.
	UpdatedAt *time.Time
	// Restore is a version restored: the text it replaces is kept even within versionEvery, it may be nowhere else.
	Restore bool
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
// TODO: the read is not locked, two edits within the same instant can both pass the UpdatedAt check; SELECT … FOR UPDATE if that happens.
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
	// A note never edited has no updated_at: its version is created_at.
	seen := p.cache.note.Item.UpdatedAt
	if seen.IsZero() {
		seen = p.cache.note.Item.CreatedAt
	}
	if p.input.UpdatedAt != nil && !seen.Equal(*p.input.UpdatedAt) {
		return ErrStale
	}

	input := editnote.Input{ID: p.input.ID, Content: p.cache.note.Content}
	if p.input.Content != nil {
		input.Content = *p.input.Content
	}
	input.Tags = hashtag.Parse(input.Content)
	if p.input.Tags != nil {
		input.Tags = *p.input.Tags
	}

	if err := p.keepVersion(tx); err != nil {
		return err
	}

	process := editnote.Process(p.ctx, tx, input)
	process.Run()
	if process.HasErr() {
		return process.GetErr()
	}
	// The saved text is read back: editnote has turned its [[links]] into note ids.
	return p.readNote(tx)
}

// keepVersion copies the text about to be replaced into note_versions, unless a version is younger than versionEvery
// (a restore keeps it anyway), then drops the versions past the versionsKept last.
func (p *run) keepVersion(tx bun.Tx) error {
	old := p.cache.note.Content
	if p.input.Content == nil || *p.input.Content == old || old == "" {
		return nil
	}

	id := p.cache.note.Item.ID
	recent, err := tx.NewSelect().
		Model((*domain.NoteVersion)(nil)).
		Where("item_id = ? AND created_at > current_timestamp - ?::interval", id, versionEvery).
		Exists(p.ctx)
	if err != nil || recent && !p.input.Restore {
		return err
	}

	content, err := secret.Seal(domain.UserID(p.ctx), old)
	if err != nil {
		return err
	}
	version := domain.NoteVersion{ItemID: id, Content: content}
	for _, match := range domain.FilePattern.FindAllStringSubmatch(old, -1) {
		version.FileIDs = append(version.FileIDs, match[1])
	}
	_, err = tx.NewInsert().Model(&version).Exec(p.ctx)
	if err != nil {
		return err
	}
	_, err = tx.NewDelete().
		Model((*domain.NoteVersion)(nil)).
		Where("item_id = ?", id).
		Where("id NOT IN (SELECT id FROM note_versions WHERE item_id = ? ORDER BY created_at DESC LIMIT ?)", id, versionsKept).
		Exec(p.ctx)
	return err
}
