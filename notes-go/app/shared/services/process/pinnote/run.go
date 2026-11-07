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
	ErrNotFound = errors.New("item not found")
)

type run struct {
	plumprimitive.Process

	ctx    context.Context
	db     bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	kind   domain.Kind
	id     string
	pinned bool
}

// Process pins a note or makes a folder a favorite, or undoes it, without touching updated_at: nothing in it changed.
func Process(ctx context.Context, db bun.IDB, kind domain.Kind, id string, pinned bool) *run {
	return &run{
		ctx:    ctx,
		db:     db,
		kind:   kind,
		id:     strings.TrimSpace(id),
		pinned: pinned,
	}
}

// Run matches a live item of the kind only; id::text, so an id that is not a uuid is not found rather than a SQL error.
func (p *run) Run() {
	if p.id == "" {
		p.SetErr(ErrNoValue)
		return
	}

	table := "metadata_notes"
	if p.kind == domain.FOLDER {
		table = "metadata_folders"
	}
	result, err := p.db.NewRaw(`
		UPDATE ? AS m SET pinned = ?
		FROM items AS i
		WHERE i.id = m.item_id AND i.id::text = ? AND i.kind = ? AND i.deleted_at IS NULL`,
		bun.Ident(table), p.pinned, p.id, p.kind).
		Exec(p.ctx)
	if err != nil {
		p.SetErr(err)
		return
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		p.SetErr(fmt.Errorf("%w: %s %q", ErrNotFound, p.kind, p.id))
	}
}
