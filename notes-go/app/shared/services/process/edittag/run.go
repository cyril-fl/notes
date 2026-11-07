package edittag

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/hashtag"
)

var (
	ErrNoValue       = errors.New("no tag value given")
	ErrNotFound      = errors.New("tag not found")
	ErrAlreadyExists = errors.New("tag already exists")
)

// uniqueViolation is the Postgres SQLSTATE raised here by ux_tags_value_lower.
const uniqueViolation = "23505"

type run struct {
	plumprimitive.Process
	cache

	ctx  context.Context
	db   bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	from string
	to   string
}

type cache struct {
	from   string
	to     string
	result domain.Tag
}

func Process(ctx context.Context, db bun.IDB, from string, to string) *run {
	return &run{
		ctx:  ctx,
		db:   db,
		from: from,
		to:   to,
	}
}

func (p *run) Run() {
	p.makeValues()
	p.renameInTx()
}

// Getters & Setters

// GetResult returns the renamed tag.
func (p *run) GetResult() domain.Tag {
	return p.cache.result
}

// Methods
func (p *run) makeValues() {
	p.cache.from = strings.TrimSpace(p.from)
	p.cache.to = strings.TrimSpace(p.to)

	switch {
	case p.cache.from == "" || p.cache.to == "":
		p.SetErr(ErrNoValue)
	case !hashtag.Valid(p.cache.to):
		p.SetErr(fmt.Errorf("%w: %q", hashtag.ErrInvalid, p.cache.to))
	}
}

// renameInTx renames the tag and its notes' text together: a failure changes neither.
func (p *run) renameInTx() {
	if p.HasErr() {
		return
	}

	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if err := p.updateTag(ctx, tx); err != nil {
			return err
		}
		return p.rewriteNotes(ctx, tx)
	}))
}

// updateTag matches with the same lower() as ux_tags_value_lower.
// A case-only rename (go -> Go) passes: the row does not conflict with itself.
func (p *run) updateTag(ctx context.Context, tx bun.Tx) error {
	err := tx.NewUpdate().
		Model((*domain.Tag)(nil)).
		Set("value = ?", p.cache.to).
		Where("user_id = ? AND lower(value) = lower(?)", domain.UserID(p.ctx), p.cache.from).
		Returning("*").
		Scan(ctx, &p.cache.result)

	return toDomainErr(err, p.cache.from, p.cache.to)
}

// rewriteNotes also reaches the trash: a note restored later must not bring the old tag back.
func (p *run) rewriteNotes(ctx context.Context, tx bun.Tx) error {
	var notes []domain.MetadataNote
	err := tx.NewSelect().
		Model(&notes).
		Column("item_id", "content").
		Join("JOIN notes_on_tags AS nt ON nt.note_id = mn.item_id").
		Where("nt.tag_id = ?", p.cache.result.ID).
		Scan(ctx)
	if err != nil {
		return err
	}

	for i := range notes {
		notes[i].Content = hashtag.Rename(notes[i].Content, p.cache.from, p.cache.to)
		if _, err := tx.NewUpdate().Model(&notes[i]).Column("content").WherePK().Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Helpers

// toDomainErr turns driver errors into the process errors callers can match with errors.Is.
func toDomainErr(err error, from string, to string) error {
	var pgErr pgdriver.Error

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return fmt.Errorf("%w: %q", ErrNotFound, from)
	case errors.As(err, &pgErr) && pgErr.Field('C') == uniqueViolation:
		return fmt.Errorf("%w: %q", ErrAlreadyExists, to)
	}

	return err
}
