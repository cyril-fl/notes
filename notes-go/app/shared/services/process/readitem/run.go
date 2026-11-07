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
	"main.go/app/shared/services/secret"
)

// invalidText is the Postgres SQLSTATE raised when the id is not a uuid.
const invalidText = "22P02"

var (
	ErrNoValue  = errors.New("no id given")
	ErrNotFound = errors.New("item not found")
)

// Result is the item and its path, with its content and tag values for a note, its direct children for a folder.
type Result struct {
	Item     domain.Item   `json:"item"`
	Path     string        `json:"path"`
	Content  string        `json:"content,omitempty" binding:"optional"`
	Pinned   bool          `json:"pinned,omitempty" binding:"optional"`
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
		Where("id = ? AND kind = ? AND user_id = ?", p.cache.id, p.kind, domain.UserID(p.ctx)).
		Scan(p.ctx)
	p.SetErr(toDomainErr(err, p.cache.id))
}

func (p *run) loadNote() {
	if p.HasErr() || p.kind != domain.NOTE {
		return
	}

	var content secret.Text
	err := p.db.NewSelect().
		Model((*domain.MetadataNote)(nil)).
		Column("content", "pinned").
		Where("item_id = ?", p.cache.id).
		Scan(p.ctx, &content, &p.cache.result.Pinned)
	p.cache.result.Content = string(content)
	if err != nil {
		p.SetErr(err)
		return
	}

	var tags []string
	err = p.db.NewSelect().
		Model((*domain.Tag)(nil)).
		Column("t.value").
		Join("JOIN notes_on_tags AS nt ON nt.tag_id = t.id").
		Where("nt.note_id = ?", p.cache.id).
		Scan(p.ctx, &tags)
	if err != nil {
		p.SetErr(err)
		return
	}
	p.cache.result.Tags, err = domain.OpenTags(tags)
	p.SetErr(err)
}

func (p *run) loadPath() {
	if p.HasErr() {
		return
	}

	paths, err := domain.Paths(p.ctx, p.db)
	p.cache.result.Path = paths[p.cache.result.Item.ID]
	if err == nil && p.cache.result.Path == "" {
		err = fmt.Errorf("%w: %q", ErrNotFound, p.cache.id) // in a trashed folder
	}
	p.SetErr(err)
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
