package deletetag

import (
	"context"
	"errors"
	"strings"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/hashtag"
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
	}
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
			WHERE t.user_id = ? AND lower(t.value) IN (SELECT lower(v) FROM unnest(?::text[]) AS v))`, domain.UserID(p.ctx), pgdialect.Array(p.cache.values)).
		Scan(ctx)
	if err != nil {
		return err
	}

	for i := range notes {
		notes[i].Content = hashtag.Remove(notes[i].Content, p.cache.values...)
		if _, err := tx.NewUpdate().Model(&notes[i]).Column("content").WherePK().Exec(ctx); err != nil {
			return err
		}
		if notes[i].Content == "" {
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

// deleteTags lowers both sides in SQL so matching uses the same lower() as ux_tags_value_lower.
// notes_on_tags rows go with it through ON DELETE CASCADE.
func (p *run) deleteTags(ctx context.Context, tx bun.Tx) error {
	return tx.NewDelete().
		Model((*domain.Tag)(nil)).
		Where("user_id = ? AND lower(value) IN (SELECT lower(v) FROM unnest(?::text[]) AS v)", domain.UserID(p.ctx), pgdialect.Array(p.cache.values)).
		Returning("*").
		Scan(ctx, &p.cache.result)
}
