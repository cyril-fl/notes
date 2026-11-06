package pinnote

import (
	"context"
	"errors"
	"fmt"
	"strings"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
)

var (
	ErrNoValue  = errors.New("no id given")
	ErrNotFound = errors.New("note not found")
)

type run struct {
	plumprimitive.Process

	ctx    context.Context
	db     bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	id     string
	pinned bool
}

// Process pins a note, or unpins it, without touching updated_at: its text did not change.
func Process(ctx context.Context, db bun.IDB, id string, pinned bool) *run {
	return &run{
		ctx:    ctx,
		db:     db,
		id:     strings.TrimSpace(id),
		pinned: pinned,
	}
}

// Run matches a live note only; id::text, so an id that is not a uuid is not found rather than a SQL error.
func (p *run) Run() {
	if p.id == "" {
		p.SetErr(ErrNoValue)
		return
	}

	result, err := p.db.NewRaw(`
		UPDATE metadata_notes AS mn SET pinned = ?
		FROM items AS i
		WHERE i.id = mn.item_id AND i.id::text = ? AND i.kind = ? AND i.deleted_at IS NULL`,
		p.pinned, p.id, domain.NOTE).
		Exec(p.ctx)
	if err != nil {
		p.SetErr(err)
		return
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		p.SetErr(fmt.Errorf("%w: %q", ErrNotFound, p.id))
	}
}
