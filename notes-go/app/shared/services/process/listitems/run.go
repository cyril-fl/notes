package listitems

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"
)

// invalidText is the Postgres SQLSTATE raised when the id is not a uuid.
const invalidText = "22P02"

var (
	ErrInvalidID   = errors.New("not a uuid")
	ErrInvalidKind = errors.New("invalid kind")
)

type run struct {
	plumprimitive.Process
	cache

	ctx      context.Context
	db       bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	parentID string
	kind     domain.Kind
}

type cache struct {
	parentID string
	result   []domain.Item
}

// Process lists the direct children of a folder; an empty parentID lists the root, an empty kind both kinds.
func Process(ctx context.Context, db bun.IDB, parentID string, kind domain.Kind) *run {
	return &run{
		ctx:      ctx,
		db:       db,
		parentID: parentID,
		kind:     kind,
	}
}

func (p *run) Run() {
	p.checkKind()
	p.makeParentID()
	p.loadItems()
}

// Getters & Setters

func (p *run) GetResult() []domain.Item {
	return p.cache.result
}

// Methods

func (p *run) checkKind() {
	if p.kind != "" && p.kind != domain.FOLDER && p.kind != domain.NOTE {
		p.SetErr(fmt.Errorf("%w: %s", ErrInvalidKind, p.kind))
	}
}

func (p *run) makeParentID() {
	p.cache.parentID = strings.TrimSpace(p.parentID)
}

// loadItems lists folders first, then notes, by title (sealed: ordered here); soft_delete hides the trashed ones.
// readonly marks the locked notes, for their padlock in the folder; has_folders tells the sidebar a folder opens
// without reading its children.
func (p *run) loadItems() {
	if p.HasErr() {
		return
	}

	query := p.db.NewSelect().
		Model(&p.cache.result).
		Column("id", "parent_id", "created_at", "updated_at", "kind", "title").
		ColumnExpr("COALESCE((SELECT mn.readonly FROM metadata_notes AS mn WHERE mn.item_id = i.id), false) AS readonly").
		ColumnExpr("EXISTS (SELECT 1 FROM items AS c WHERE c.parent_id = i.id AND c.kind = ? AND c.deleted_at IS NULL) AS has_folders", domain.FOLDER).
		Where("user_id = ?", domain.UserID(p.ctx)).
		OrderExpr("kind = ? DESC, id", domain.FOLDER)

	if p.cache.parentID == "" {
		query.Where("parent_id IS NULL")
	} else {
		query.Where("parent_id = ?", p.cache.parentID)
	}
	if p.kind != "" {
		query.Where("kind = ?", p.kind)
	}

	p.SetErr(toDomainErr(query.Scan(p.ctx)))
	domain.SortFold(p.cache.result, func(i domain.Item) string { return string(i.Title) }, func(a, b domain.Item) int {
		return boolOrder(b.Kind == domain.FOLDER) - boolOrder(a.Kind == domain.FOLDER)
	})
}

// Helpers

// toDomainErr turns a parent id that is not a uuid into ErrInvalidID.
func toDomainErr(err error) error {
	var pgErr pgdriver.Error
	if errors.As(err, &pgErr) && pgErr.Field('C') == invalidText {
		return fmt.Errorf("%w: %s", ErrInvalidID, pgErr.Field('M'))
	}
	return err
}

func boolOrder(b bool) int {
	if b {
		return 1
	}
	return 0
}
