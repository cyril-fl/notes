package countnotes

import (
	"context"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

// Stats is the number of live notes of the account and the bytes its notes take in the base.
type Stats struct {
	Notes int   `json:"notes"`
	Size  int64 `json:"size"` // items, texts and versions, the trash included: what the base keeps
}

type run struct {
	plumprimitive.Process
	cache

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
}

type cache struct {
	result Stats
}

func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

func (p *run) Run() {
	p.countNotes()
}

// Getters & Setters

func (p *run) GetResult() Stats {
	return p.cache.result
}

// Methods

// countNotes sums pg_column_size of the rows, so a TOASTed text counts compressed, as stored.
func (p *run) countNotes() {
	userID := domain.UserID(p.ctx)
	p.SetErr(p.db.NewRaw(`
		SELECT
			count(*) FILTER (WHERE i.kind = ? AND i.deleted_at IS NULL) AS notes,
			coalesce(sum(pg_column_size(i.*)), 0) + coalesce(sum(pg_column_size(mn.*)), 0) + (
				SELECT coalesce(sum(pg_column_size(nv.*)), 0)
				FROM note_versions AS nv JOIN items AS vi ON vi.id = nv.item_id
				WHERE vi.user_id = ?
			) AS size
		FROM items AS i
		LEFT JOIN metadata_notes AS mn ON mn.item_id = i.id
		WHERE i.user_id = ?`, domain.NOTE, userID, userID).
		Scan(p.ctx, &p.cache.result))
}
