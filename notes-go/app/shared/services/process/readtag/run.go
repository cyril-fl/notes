package readtag

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"main.go/app/shared/services/secret"
	"slices"
	"strings"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
)

var (
	ErrNoValue  = errors.New("no tag value given")
	ErrNotFound = errors.New("tag not found")
)

// Note is a note carrying the tag, with its path.
type Note struct {
	ID    string      `json:"id"`
	Title secret.Text `json:"title" swaggertype:"string"`
	Path  string      `json:"path"`
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

// loadTag matches by the blind of ux_tags_user_value_hash, whatever the case.
func (p *run) loadTag() {
	if p.HasErr() {
		return
	}

	hash, err := secret.Blind(domain.UserID(p.ctx), p.cache.value)
	if err != nil {
		p.SetErr(err)
		return
	}
	err = p.db.NewSelect().
		Model(&p.cache.result.Tag).
		Where("user_id = ? AND value_hash = ?", domain.UserID(p.ctx), hash).
		Scan(p.ctx)
	if errors.Is(err, sql.ErrNoRows) {
		err = fmt.Errorf("%w: %q", ErrNotFound, p.cache.value)
	}
	p.SetErr(err)
}

// loadNotes joins on live, which only holds live items: trashed notes drop out.
func (p *run) loadNotes() {
	if p.HasErr() {
		return
	}

	p.SetErr(p.db.NewRaw(domain.LiveCTE+`
		SELECT i.id, i.title
		FROM notes_on_tags AS nt
		JOIN items AS i ON i.id = nt.note_id
		JOIN live ON live.id = i.id
		WHERE nt.tag_id = ?`, domain.UserID(p.ctx), p.cache.result.Tag.ID).
		Scan(p.ctx, &p.cache.result.Notes))
	if p.HasErr() {
		return
	}

	paths, err := domain.Paths(p.ctx, p.db)
	p.SetErr(err)
	for i := range p.cache.result.Notes {
		p.cache.result.Notes[i].Path = paths[p.cache.result.Notes[i].ID]
	}
	slices.SortFunc(p.cache.result.Notes, func(a, b Note) int { return domain.CompareFold(a.Path, b.Path) })
}
