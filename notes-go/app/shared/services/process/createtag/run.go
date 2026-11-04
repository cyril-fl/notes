package createtag

import (
	"context"
	"errors"
	"fmt"
	"strings"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
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
	tags   []domain.Tag
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
	p.makeTags()
	p.insertTags()
}

// Getters & Setters

// GetResult returns only the tags actually created, existing ones are skipped.
func (p *run) GetResult() []domain.Tag {
	return p.cache.result
}

// Methods
func (p *run) makeTags() {
	for _, value := range p.values {
		if value = strings.TrimSpace(value); value == "" {
			continue
		}
		if !hashtag.Valid(value) {
			p.SetErr(fmt.Errorf("%w: %q", hashtag.ErrInvalid, value))
			return
		}
		p.cache.tags = append(p.cache.tags, domain.Tag{Value: value})
	}

	if len(p.cache.tags) == 0 {
		p.SetErr(ErrNoValue)
	}
}

// insertTags scans into result, not tags: RETURNING only yields inserted rows
// and bun would overwrite the first elements of tags in order.
// Conflict target must match the ux_tags_value_lower expression index.
func (p *run) insertTags() {
	if p.HasErr() {
		return
	}

	p.SetErr(p.db.NewInsert().
		Model(&p.cache.tags).
		On("CONFLICT ((lower(value))) DO NOTHING").
		Returning("*").
		Scan(p.ctx, &p.cache.result))
}
