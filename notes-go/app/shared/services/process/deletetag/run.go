package deletetag

import (
	"context"
	"errors"
	"strings"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"main.go/app/shared/domain"
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
	p.deleteTags()
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

// deleteTags lowers both sides in SQL so matching uses the same lower() as ux_tags_value_lower.
// notes_on_tags rows go with it through ON DELETE CASCADE, notes are untouched.
func (p *run) deleteTags() {
	if p.HasErr() {
		return
	}

	p.SetErr(p.db.NewDelete().
		Model((*domain.Tag)(nil)).
		Where("lower(value) IN (SELECT lower(v) FROM unnest(?::text[]) AS v)", pgdialect.Array(p.cache.values)).
		Returning("*").
		Scan(p.ctx, &p.cache.result))
}
