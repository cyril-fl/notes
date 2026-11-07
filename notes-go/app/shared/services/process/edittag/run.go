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
	"main.go/app/shared/services/secret"
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

// updateTag matches by the blind of ux_tags_user_value_hash, whatever the case.
// A case-only rename (go -> Go) passes: the row does not conflict with itself.
func (p *run) updateTag(ctx context.Context, tx bun.Tx) error {
	from, err := secret.Blind(domain.UserID(p.ctx), p.cache.from)
	if err != nil {
		return err
	}
	to, err := domain.SealTag(p.ctx, p.cache.to)
	if err != nil {
		return err
	}
	err = tx.NewUpdate().
		Model((*domain.Tag)(nil)).
		Set("value = ?, value_hash = ?", to.Value, to.ValueHash).
		Where("user_id = ? AND value_hash = ?", domain.UserID(p.ctx), from).
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
		sealed, err := secret.Seal(domain.UserID(p.ctx), hashtag.Rename(string(notes[i].Content), p.cache.from, p.cache.to))
		if err != nil {
			return err
		}
		if _, err := tx.NewUpdate().Model((*domain.MetadataNote)(nil)).Set("content = ?", sealed).Where("item_id = ?", notes[i].ItemID).Exec(ctx); err != nil {
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
