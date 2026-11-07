package listitems

import (
	"context"
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

var ErrInvalidID = errors.New("not a uuid")

type run struct {
	plumprimitive.Process
	cache

	ctx      context.Context
	db       bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	parentID string
}

type cache struct {
	parentID string
	result   []domain.Item
}

// Process lists the direct children of a folder; an empty parentID lists the root.
func Process(ctx context.Context, db bun.IDB, parentID string) *run {
	return &run{
		ctx:      ctx,
		db:       db,
		parentID: parentID,
	}
}

func (p *run) Run() {
	p.makeParentID()
	p.loadItems()
}

// Getters & Setters

func (p *run) GetResult() []domain.Item {
	return p.cache.result
}

// Methods

func (p *run) makeParentID() {
	p.cache.parentID = strings.TrimSpace(p.parentID)
}

// loadItems lists folders first, then notes, by title; soft_delete hides the trashed ones.
func (p *run) loadItems() {
	query := p.db.NewSelect().
		Model(&p.cache.result).
		Column("id", "parent_id", "created_at", "updated_at", "kind", "title").
		Where("user_id = ?", domain.UserID(p.ctx)).
		OrderExpr("kind = ? DESC, lower(title)", domain.FOLDER)

	if p.cache.parentID == "" {
		query.Where("parent_id IS NULL")
	} else {
		query.Where("parent_id = ?", p.cache.parentID)
	}

	p.SetErr(toDomainErr(query.Scan(p.ctx)))
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
