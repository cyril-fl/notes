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
	"main.go/app/shared/services/process/listitems"
)

// invalidText is the Postgres SQLSTATE raised when the id is not a uuid.
const invalidText = "22P02"

// PathsCTE gives every live item its path from the root (/Projects/Go/Ideas): append a query on "paths".
// It walks down from the roots, so a trashed folder hides its whole subtree.
const PathsCTE = `WITH RECURSIVE paths AS (
	SELECT id, '/' || coalesce(title, 'untitled') AS path FROM items WHERE parent_id IS NULL AND deleted_at IS NULL
	UNION ALL SELECT child.id, paths.path || '/' || coalesce(child.title, 'untitled')
	FROM items AS child JOIN paths ON child.parent_id = paths.id WHERE child.deleted_at IS NULL
) `

var (
	ErrNoValue  = errors.New("no id given")
	ErrNotFound = errors.New("item not found")
)

// Result is the item and its path, with its content and tag values for a note, its direct children for a folder.
type Result struct {
	Item     domain.Item   `json:"item"`
	Path     string        `json:"path"`
	Content  string        `json:"content,omitempty" binding:"optional"`
	Tags     []string      `json:"tags,omitzero" binding:"optional"`
	Children []domain.Item `json:"children,omitzero" binding:"optional"`
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
	p.loadPath()
	p.loadNote()
	p.loadChildren()
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

func (p *run) loadPath() {
	if p.HasErr() {
		return
	}

	p.SetErr(p.db.NewRaw(PathsCTE+"SELECT path FROM paths WHERE id = ?", p.cache.id).Scan(p.ctx, &p.cache.result.Path))
}

func (p *run) loadChildren() {
	if p.HasErr() || p.kind != domain.FOLDER {
		return
	}

	process := listitems.Process(p.ctx, p.db, p.cache.id)
	process.Run()
	p.cache.result.Children = process.GetResult()
	p.SetErr(process.GetErr())
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
