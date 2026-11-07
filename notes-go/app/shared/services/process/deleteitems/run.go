package deleteitems

import (
	"context"
	"errors"
	"strings"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/process/linknotes"
)

var ErrNoValue = errors.New("no id given")

type run struct {
	plumprimitive.Process
	cache

	ctx  context.Context
	db   bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	kind domain.Kind
	ids  []string
}

type cache struct {
	ids    []string
	result []domain.Item
}

// Process moves the items of the given kind to the trash; a folder takes its whole subtree with it.
func Process(ctx context.Context, db bun.IDB, kind domain.Kind, ids []string) *run {
	return &run{
		ctx:  ctx,
		db:   db,
		kind: kind,
		ids:  ids,
	}
}

func (p *run) Run() {
	p.makeIDs()
	p.markDeleted()
}

// Getters & Setters

// GetResult returns every item marked: the given ones and their descendants.
func (p *run) GetResult() []domain.Item {
	return p.cache.result
}

// Methods

func (p *run) makeIDs() {
	for _, id := range p.ids {
		if id = strings.TrimSpace(id); id != "" {
			p.cache.ids = append(p.cache.ids, id)
		}
	}

	if len(p.cache.ids) == 0 {
		p.SetErr(ErrNoValue)
	}
}

// markDeleted is one statement: current_timestamp is the same for the whole subtree, so a restore can find it back.
// Notes never have children, so the recursion is a no-op for them.
// Unknown, already deleted, or other-kind ids match nothing and are skipped.
// The [[links]] waiting for a trashed note's title are resolved again in the same transaction: a title it shared
// may now name one note only. A link to it stays, hidden by the readers until a restore.
func (p *run) markDeleted() {
	if p.HasErr() {
		return
	}

	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		err := tx.NewUpdate().
			Model((*domain.Item)(nil)).
			Set("deleted_at = current_timestamp").
			Where(`id IN (
				WITH RECURSIVE tree AS (
					SELECT id FROM items WHERE kind = ? AND id IN (?) AND user_id = ? AND deleted_at IS NULL
					UNION SELECT child.id FROM items AS child JOIN tree ON child.parent_id = tree.id
				) SELECT id FROM tree)`, p.kind, bun.In(p.cache.ids), domain.UserID(p.ctx)).
			Returning("id, kind, title").
			Scan(ctx, &p.cache.result)
		if err != nil {
			return err
		}
		return linknotes.Items(ctx, tx, p.cache.result)
	}))
}
