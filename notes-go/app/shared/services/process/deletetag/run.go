package deletetag

import (
	"context"
	"errors"
	"strings"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/hashtag"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

var ErrNoValue = errors.New("no tag value given")

type run struct {
	plumprimitive.Process
	cache

	ctx    context.Context
	db     bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	values []string
}

type cache struct {
	values []string
	hashes []string // the value_hash of values
	result []domain.Tag
}

func Process(ctx context.Context, db bun.IDB, values []string) *run {
	return &run{
		ctx:    ctx,
		db:     db,
		values: values,
	}
}

func (p *run) Run() {
	p.makeValues()
	p.deleteInTx()
}

// Getters & Setters

// GetResult returns only the tags actually deleted, unknown values are skipped.
func (p *run) GetResult() []domain.Tag {
	return p.cache.result
}

// Methods
func (p *run) makeValues() {
	for _, value := range p.values {
		if value = strings.TrimSpace(value); value != "" {
			p.cache.values = append(p.cache.values, value)
		}
	}

	if len(p.cache.values) == 0 {
		p.SetErr(ErrNoValue)
		return
	}
	var err error
	p.cache.hashes, err = domain.BlindTags(p.ctx, p.cache.values)
	p.SetErr(err)
}

// deleteInTx erases the tags from their notes' text and deletes them together: a failure changes neither.
func (p *run) deleteInTx() {
	if p.HasErr() {
		return
	}

	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if err := p.rewriteNotes(ctx, tx); err != nil {
			return err
		}
		return p.deleteTags(ctx, tx)
	}))
}

// rewriteNotes also reaches the trash: a note restored later must not bring the tag back.
func (p *run) rewriteNotes(ctx context.Context, tx bun.Tx) error {
	var notes []domain.MetadataNote
	err := tx.NewSelect().
		Model(&notes).
		Column("item_id", "content").
		Where(`item_id IN (
			SELECT nt.note_id FROM notes_on_tags AS nt JOIN tags AS t ON t.id = nt.tag_id
			WHERE t.user_id = ? AND t.value_hash IN (?))`, domain.UserID(p.ctx), bun.In(p.cache.hashes)).
		Scan(ctx)
	if err != nil {
		return err
	}

	for i := range notes {
		content := hashtag.Remove(string(notes[i].Content), p.cache.values...)
		sealed, err := secret.Seal(domain.UserID(p.ctx), content)
		if err != nil {
			return err
		}
		if _, err := tx.NewUpdate().Model((*domain.MetadataNote)(nil)).Set("content = ?", sealed).Where("item_id = ?", notes[i].ItemID).Exec(ctx); err != nil {
			return err
		}
		if content == "" {
			if err := trashNote(ctx, tx, notes[i].ItemID); err != nil {
				return err
			}
		}
	}
	return nil
}

// trashNote is for a note that held nothing but the tag: an empty note goes to the trash.
func trashNote(ctx context.Context, tx bun.Tx, id string) error {
	_, err := tx.NewUpdate().
		Model((*domain.Item)(nil)).
		Set("deleted_at = current_timestamp").
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// deleteTags matches by the blinds of ux_tags_user_value_hash, whatever the case.
// notes_on_tags rows go with it through ON DELETE CASCADE.
func (p *run) deleteTags(ctx context.Context, tx bun.Tx) error {
	return tx.NewDelete().
		Model((*domain.Tag)(nil)).
		Where("user_id = ? AND value_hash IN (?)", domain.UserID(p.ctx), bun.In(p.cache.hashes)).
		Returning("?Columns").
		Scan(ctx, &p.cache.result)
}
