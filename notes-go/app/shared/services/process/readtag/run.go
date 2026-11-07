package readtag

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/process/readitem"
)

var (
	ErrNoValue  = errors.New("no tag value given")
	ErrNotFound = errors.New("tag not found")
)

// Note is a note carrying the tag, with its path.
type Note struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

// Result is the tag and its live notes, by path.
type Result struct {
	Tag   domain.Tag `json:"tag"`
	Notes []Note     `json:"notes"`
}

type run struct {
	plumprimitive.Process
	cache

	ctx   context.Context
	db    bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	value string
}

type cache struct {
	value  string
	result Result
}

func Process(ctx context.Context, db bun.IDB, value string) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		value: value,
	}
}

func (p *run) Run() {
	p.makeValue()
	p.loadTag()
	p.loadNotes()
}

// Getters & Setters

func (p *run) GetResult() Result {
	return p.cache.result
}

// Methods

func (p *run) makeValue() {
	p.cache.value = strings.TrimSpace(p.value)
	if p.cache.value == "" {
		p.SetErr(ErrNoValue)
	}
}

// loadTag matches with the same lower() as ux_tags_value_lower.
func (p *run) loadTag() {
	if p.HasErr() {
		return
	}

	err := p.db.NewSelect().
		Model(&p.cache.result.Tag).
		Where("user_id = ? AND lower(value) = lower(?)", domain.UserID(p.ctx), p.cache.value).
		Scan(p.ctx)
	if errors.Is(err, sql.ErrNoRows) {
		err = fmt.Errorf("%w: %q", ErrNotFound, p.cache.value)
	}
	p.SetErr(err)
}

// loadNotes joins on paths, which only holds live items: trashed notes drop out.
func (p *run) loadNotes() {
	if p.HasErr() {
		return
	}

	p.SetErr(p.db.NewRaw(readitem.PathsCTE+`
		SELECT i.id, coalesce(i.title, '') AS title, paths.path
		FROM notes_on_tags AS nt
		JOIN items AS i ON i.id = nt.note_id
		JOIN paths ON paths.id = i.id
		WHERE nt.tag_id = ?
		ORDER BY lower(paths.path)`, domain.UserID(p.ctx), p.cache.result.Tag.ID).
		Scan(p.ctx, &p.cache.result.Notes))
}
