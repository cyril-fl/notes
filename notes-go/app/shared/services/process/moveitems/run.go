package moveitems

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

// invalidText is the Postgres SQLSTATE raised when an id is not a uuid.
const invalidText = "22P02"

var (
	ErrNoValue        = errors.New("no id given")
	ErrInvalidID      = errors.New("not a uuid")
	ErrParentNotFound = errors.New("parent folder not found")
	ErrCycle          = errors.New("a folder cannot move into itself or its subfolders")
)

// Input lists the items to move: an empty ParentID moves them to the root.
type Input struct {
	IDs      []string
	ParentID string
}

type run struct {
	plumprimitive.Process
	cache

	ctx   context.Context
	db    bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	input Input
}

type cache struct {
	input  Input
	result []domain.Item
}

func Process(ctx context.Context, db bun.IDB, input Input) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		input: input,
	}
}

func (p *run) Run() {
	p.makeInput()
	p.moveItems()
}

// Getters & Setters

// GetResult returns the moved items.
func (p *run) GetResult() []domain.Item {
	return p.cache.result
}

// Methods

func (p *run) makeInput() {
	p.cache.input.ParentID = strings.TrimSpace(p.input.ParentID)
	for _, id := range p.input.IDs {
		if id = strings.TrimSpace(id); id != "" {
			p.cache.input.IDs = append(p.cache.input.IDs, id)
		}
	}

	if len(p.cache.input.IDs) == 0 {
		p.SetErr(ErrNoValue)
	}
}

// moveItems checks the parent then updates in one transaction, so the parent cannot change in between.
func (p *run) moveItems() {
	if p.HasErr() {
		return
	}

	err := p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, step := range []func(context.Context, bun.Tx) error{p.checkParent, p.checkCycle, p.updateItems} {
			if err := step(ctx, tx); err != nil {
				return err
			}
		}
		return nil
	})
	p.SetErr(toDomainErr(err))
}

// checkParent wants a folder not in the trash: the FK alone accepts a trashed one.
func (p *run) checkParent(ctx context.Context, tx bun.Tx) error {
	if p.cache.input.ParentID == "" {
		return nil
	}

	exists, err := tx.NewSelect().
		Model((*domain.Item)(nil)).
		Where("id = ? AND kind = ?", p.cache.input.ParentID, domain.FOLDER).
		Exists(ctx)
	if err == nil && !exists {
		err = fmt.Errorf("%w: %q", ErrParentNotFound, p.cache.input.ParentID)
	}
	return err
}

// checkCycle walks up from the parent: a moved item met on the way is the parent or one of its ancestors.
func (p *run) checkCycle(ctx context.Context, tx bun.Tx) error {
	if p.cache.input.ParentID == "" {
		return nil
	}

	var cycle bool
	err := tx.NewRaw(`
		WITH RECURSIVE up AS (
			SELECT id, parent_id FROM items WHERE id = ?
			UNION SELECT parent.id, parent.parent_id FROM items AS parent JOIN up ON parent.id = up.parent_id
		) SELECT EXISTS (SELECT 1 FROM up WHERE id IN (?))`,
		p.cache.input.ParentID, bun.In(p.cache.input.IDs)).
		Scan(ctx, &cycle)
	if err == nil && cycle {
		err = ErrCycle
	}
	return err
}

// updateItems is one statement; unknown or trashed ids match nothing and are skipped (soft_delete adds deleted_at IS NULL).
func (p *run) updateItems(ctx context.Context, tx bun.Tx) error {
	var parentID *string // nil = NULL = root
	if p.cache.input.ParentID != "" {
		parentID = &p.cache.input.ParentID
	}

	return tx.NewUpdate().
		Model((*domain.Item)(nil)).
		Set("parent_id = ?", parentID).
		Set("updated_at = current_timestamp").
		Where("id IN (?)", bun.In(p.cache.input.IDs)).
		Returning("id, parent_id, kind, title").
		Scan(ctx, &p.cache.result)
}

// Helpers

// toDomainErr turns a value that is not a uuid into ErrInvalidID.
func toDomainErr(err error) error {
	var pgErr pgdriver.Error
	if errors.As(err, &pgErr) && pgErr.Field('C') == invalidText {
		return fmt.Errorf("%w: %s", ErrInvalidID, pgErr.Field('M'))
	}
	return err
}
