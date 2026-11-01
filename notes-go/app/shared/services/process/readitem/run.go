package readitem

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

// invalidText is the Postgres SQLSTATE raised when the id is not a uuid.
const invalidText = "22P02"

var (
	ErrNoValue  = errors.New("no id given")
	ErrNotFound = errors.New("item not found")
)

// Result is the item with, for a note, its content and tag values.
type Result struct {
	Item    domain.Item
	Content string
	Tags    []string
}

type run struct {
	plumprimitive.Process
	cache

	ctx  context.Context
	db   bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	kind domain.Kind
	id   string
}

type cache struct {
	id     string
	result Result
}

// Process loads one item of the given kind that is not in the trash.
func Process(ctx context.Context, db bun.IDB, kind domain.Kind, id string) *run {
	return &run{
		ctx:  ctx,
		db:   db,
		kind: kind,
		id:   id,
	}
}

func (p *run) Run() {
	p.makeID()
	p.loadItem()
	p.loadNote()
}

// Getters & Setters

func (p *run) GetResult() Result {
	return p.cache.result
}

// Methods

func (p *run) makeID() {
	p.cache.id = strings.TrimSpace(p.id)
	if p.cache.id == "" {
		p.SetErr(ErrNoValue)
	}
}

// loadItem relies on soft_delete: bun adds deleted_at IS NULL, a trashed item is not found.
func (p *run) loadItem() {
	if p.HasErr() {
		return
	}

	// Not "*": parent_kind is a DB-only column, Item cannot scan it.
	err := p.db.NewSelect().
		Model(&p.cache.result.Item).
		Column("id", "parent_id", "created_at", "updated_at", "kind", "title").
		Where("id = ? AND kind = ?", p.cache.id, p.kind).
		Scan(p.ctx)
	p.SetErr(toDomainErr(err, p.cache.id))
}

func (p *run) loadNote() {
	if p.HasErr() || p.kind != domain.NOTE {
		return
	}

	err := p.db.NewSelect().
		Model((*domain.MetadataNote)(nil)).
		Column("content").
		Where("item_id = ?", p.cache.id).
		Scan(p.ctx, &p.cache.result.Content)
	if err != nil {
		p.SetErr(err)
		return
	}

	p.SetErr(p.db.NewSelect().
		Model((*domain.Tag)(nil)).
		Column("t.value").
		Join("JOIN notes_on_tags AS nt ON nt.tag_id = t.id").
		Where("nt.note_id = ?", p.cache.id).
		OrderExpr("lower(t.value)").
		Scan(p.ctx, &p.cache.result.Tags))
}

// Helpers

// toDomainErr turns no row, or an id that is not a uuid, into ErrNotFound.
func toDomainErr(err error, id string) error {
	var pgErr pgdriver.Error
	if errors.Is(err, sql.ErrNoRows) || errors.As(err, &pgErr) && pgErr.Field('C') == invalidText {
		return fmt.Errorf("%w: %q", ErrNotFound, id)
	}
	return err
}
