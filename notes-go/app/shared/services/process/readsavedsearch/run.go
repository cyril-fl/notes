// Package readsavedsearch finds a saved search by id among the account's, then replays its criteria.
package readsavedsearch

import (
	"context"
	"errors"
	"fmt"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/runsavedsearch"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

var ErrNotFound = errors.New("saved search not found")

type run struct {
	plumprimitive.Process
	cache

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	id  string
}

type cache struct {
	search domain.SavedSearch
	result []runsavedsearch.Match
}

func Process(ctx context.Context, db bun.IDB, id string) *run {
	return &run{
		ctx: ctx,
		db:  db,
		id:  id,
	}
}

func (p *run) Run() {
	p.findSearch()
	p.runSearch()
}

// Getters & Setters

func (p *run) GetResult() []runsavedsearch.Match {
	return p.cache.result
}

// Methods

// findSearch loads the account's saved searches and keeps the one asked for.
func (p *run) findSearch() {
	var searches domain.SavedSearches
	err := p.db.NewRaw("SELECT saved_searches FROM users WHERE id = ?", domain.UserID(p.ctx)).Scan(p.ctx, &searches)
	if err != nil {
		p.SetErr(err)
		return
	}

	for _, search := range searches {
		if search.ID == p.id {
			p.cache.search = search
			return
		}
	}
	p.SetErr(fmt.Errorf("%w: %q", ErrNotFound, p.id))
}

func (p *run) runSearch() {
	if p.HasErr() {
		return
	}

	process := runsavedsearch.Process(p.ctx, p.db, p.cache.search)
	process.Run()
	p.SetErr(process.GetErr())
	p.cache.result = process.GetResult()
}
