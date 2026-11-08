package readfile

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/silo"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

var ErrNotFound = errors.New("file not found")

type run struct {
	plumprimitive.Process
	cache

	ctx   context.Context
	db    bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	store silo.Store
	id    string
}

type cache struct {
	file domain.File
	body io.ReadCloser
}

// Process opens a file of the account: the caller closes GetBody.
func Process(ctx context.Context, db bun.IDB, store silo.Store, id string) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		store: store,
		id:    id,
	}
}

func (p *run) Run() {
	p.loadFile()
	p.openBody()
}

// Getters & Setters

func (p *run) GetResult() domain.File {
	return p.cache.file
}

func (p *run) GetBody() io.ReadCloser {
	return p.cache.body
}

// Methods

// loadFile compares id::text: an id that is not a uuid is not found rather than a SQL error.
func (p *run) loadFile() {
	err := p.db.NewSelect().Model(&p.cache.file).Where("id::text = ? AND user_id = ?", p.id, domain.UserID(p.ctx)).Scan(p.ctx)
	if errors.Is(err, sql.ErrNoRows) {
		err = fmt.Errorf("%w: %q", ErrNotFound, p.id)
	}
	p.SetErr(err)
}

func (p *run) openBody() {
	if p.HasErr() {
		return
	}

	var err error
	p.cache.body, err = p.store.Get(p.ctx, p.cache.file.Key())
	p.SetErr(err)
}
