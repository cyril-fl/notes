package restoreitems

import (
	"context"
	"errors"
	"fmt"
	"strings"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/process/linknotes"
)

// invalidText is the Postgres SQLSTATE raised when an id is not a uuid.
const invalidText = "22P02"

var (
	ErrNoValue       = errors.New("no id given")
	ErrInvalidID     = errors.New("not a uuid")
	ErrParentInTrash = errors.New("the parent folder is in the trash: restore it first")
)

type run struct {
	plumprimitive.Process
	cache

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	ids []string
}

type cache struct {
	ids    []string
	result []domain.Item
}

// Process takes items out of the trash, with everything removed at the same time as them.
func Process(ctx context.Context, db bun.IDB, ids []string) *run {
	return &run{
		ctx: ctx,
		db:  db,
		ids: ids,
	}
}

func (p *run) Run() {
	p.makeIDs()
	p.restoreItems()
}

// Getters & Setters

// GetResult returns every item restored: the given ones and what went to the trash with them.
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

// restoreItems checks the parents then restores in one transaction. Raw SQL: bun's soft_delete would hide trashed rows.
func (p *run) restoreItems() {
	if p.HasErr() {
		return
	}

	err := p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if err := p.checkParents(ctx, tx); err != nil {
			return err
		}
		if err := p.updateItems(ctx, tx); err != nil {
			return err
		}
		return p.linkNotes(ctx, tx)
	})
	p.SetErr(toDomainErr(err))
}

// checkParents refuses an item whose parent stays in the trash: it would be live under a trashed folder,
// invisible everywhere. A parent restored in the same call is fine.
func (p *run) checkParents(ctx context.Context, tx bun.Tx) error {
	var titles []string
	err := tx.NewRaw(`
		SELECT coalesce(i.title, 'untitled')
		FROM items AS i JOIN items AS parent ON parent.id = i.parent_id
		WHERE i.id IN (?) AND i.user_id = ? AND i.deleted_at IS NOT NULL AND parent.deleted_at IS NOT NULL AND parent.id NOT IN (?)`,
		bun.In(p.cache.ids), domain.UserID(p.ctx), bun.In(p.cache.ids)).
		Scan(ctx, &titles)
	if err == nil && len(titles) > 0 {
		err = fmt.Errorf("%w: %s", ErrParentInTrash, strings.Join(titles, ", "))
	}
	return err
}

// updateItems walks down from each item through the children removed at the same time: a child removed
// earlier on its own stays in the trash. Unknown or live ids match nothing and are skipped.
func (p *run) updateItems(ctx context.Context, tx bun.Tx) error {
	return tx.NewRaw(`
		UPDATE items SET deleted_at = NULL
		WHERE id IN (
			WITH RECURSIVE tree AS (
				SELECT id, deleted_at FROM items WHERE id IN (?) AND user_id = ? AND deleted_at IS NOT NULL
				UNION SELECT child.id, child.deleted_at FROM items AS child JOIN tree ON child.parent_id = tree.id
				WHERE child.deleted_at = tree.deleted_at
			) SELECT id FROM tree)
		RETURNING id, parent_id, created_at, updated_at, kind, title`, bun.In(p.cache.ids), domain.UserID(p.ctx)).
		Scan(ctx, &p.cache.result)
}

// linkNotes links each restored note again: its own [[links]], and the [[links]] of the notes waiting for its
// title. A link to a trashed note stays in the text and in notes_on_links, hidden by the readers: it is back here.
func (p *run) linkNotes(ctx context.Context, tx bun.Tx) error {
	return linknotes.Items(ctx, tx, p.cache.result)
}

// Helpers

// toDomainErr turns a value that is not a uuid into ErrInvalidID.
func toDomainErr(err error) error {
	if domain.FolderNameTaken(err) {
		return domain.ErrFolderExists
	}
	var pgErr pgdriver.Error
	if errors.As(err, &pgErr) && pgErr.Field('C') == invalidText {
		return fmt.Errorf("%w: %s", ErrInvalidID, pgErr.Field('M'))
	}
	return err
}
