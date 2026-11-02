package listtrash

import (
	"context"
	"time"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
)

// Trashed is what was removed in one go: the item and how many items went with it.
type Trashed struct {
	ID        string      `json:"id"`
	Kind      domain.Kind `json:"kind"`
	Title     string      `json:"title"`
	DeletedAt time.Time   `json:"deleted_at"`
	Inside    int         `json:"inside"`
}

type run struct {
	plumprimitive.Process
	cache

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
}

type cache struct {
	result []Trashed
}

func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

func (p *run) Run() {
	p.loadTrash()
}

// Getters & Setters

// GetResult returns the trash, last removed first.
func (p *run) GetResult() []Trashed {
	return p.cache.result
}

// Methods

// loadTrash lists the root of each removal: deleteitems gives a whole subtree the same deleted_at, so an item whose
// parent was removed at the same time is inside that removal, not one of its own. Raw SQL: bun's soft_delete would hide them.
func (p *run) loadTrash() {
	p.SetErr(p.db.NewRaw(`
		SELECT i.id, i.kind, coalesce(i.title, '') AS title, i.deleted_at,
			(WITH RECURSIVE inside AS (
				SELECT child.id FROM items AS child WHERE child.parent_id = i.id AND child.deleted_at = i.deleted_at
				UNION SELECT child.id FROM items AS child JOIN inside ON child.parent_id = inside.id WHERE child.deleted_at = i.deleted_at
			) SELECT count(*) FROM inside) AS inside
		FROM items AS i
		LEFT JOIN items AS parent ON parent.id = i.parent_id
		WHERE i.deleted_at IS NOT NULL AND parent.deleted_at IS DISTINCT FROM i.deleted_at
		ORDER BY i.deleted_at DESC, lower(i.title)`).
		Scan(p.ctx, &p.cache.result))
}
