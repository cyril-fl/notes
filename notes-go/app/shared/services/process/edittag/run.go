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
	p.updateTag()
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

	if p.cache.from == "" || p.cache.to == "" {
		p.SetErr(ErrNoValue)
	}
}

// updateTag matches with the same lower() as ux_tags_value_lower.
// A case-only rename (go -> Go) passes: the row does not conflict with itself.
func (p *run) updateTag() {
	if p.HasErr() {
		return
	}

	err := p.db.NewUpdate().
		Model((*domain.Tag)(nil)).
		Set("value = ?", p.cache.to).
		Where("lower(value) = lower(?)", p.cache.from).
		Returning("*").
		Scan(p.ctx, &p.cache.result)

	p.SetErr(toDomainErr(err, p.cache.from, p.cache.to))
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
