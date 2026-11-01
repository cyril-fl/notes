package editfolder

import (
	"context"
	"database/sql"
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

var (
	ErrNoValue  = errors.New("no id given")
	ErrNoTitle  = errors.New("no folder title given")
	ErrNotFound = errors.New("folder not found")
)

// Input is the new state of the folder.
type Input struct {
	ID    string
	Title string
}

type run struct {
	plumprimitive.Process
	cache

	ctx   context.Context
	db    bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	input Input
}

type cache struct {
	input Input
	item  domain.Item
}

func Process(ctx context.Context, db bun.IDB, input Input) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		input: input,
	}
}

func (p *run) Run() {
	p.makeFolder()
	p.updateFolder()
}

// Getters & Setters

// GetResult returns the updated item.
func (p *run) GetResult() domain.Item {
	return p.cache.item
}

// Methods

func (p *run) makeFolder() {
	p.cache.input = Input{
		ID:    strings.TrimSpace(p.input.ID),
		Title: strings.TrimSpace(p.input.Title),
	}

	switch {
	case p.cache.input.ID == "":
		p.SetErr(ErrNoValue)
	case p.cache.input.Title == "":
		p.SetErr(ErrNoTitle)
	}
}

// updateFolder matches a folder not in the trash (soft_delete adds deleted_at IS NULL): anything else is not found.
func (p *run) updateFolder() {
	if p.HasErr() {
		return
	}

	err := p.db.NewUpdate().
		Model(&p.cache.item).
		Set("title = ?", p.cache.input.Title).
		Set("updated_at = current_timestamp").
		Where("id = ? AND kind = ?", p.cache.input.ID, domain.FOLDER).
		Returning("id, parent_id, created_at, updated_at, kind, title").
		Scan(p.ctx)
	p.SetErr(toDomainErr(err, p.cache.input.ID))
}

// Helpers

// toDomainErr turns no row, or an id that is not a uuid, into ErrNotFound.
func toDomainErr(err error, id string) error {
	var pgErr pgdriver.Error
	if errors.Is(err, sql.ErrNoRows) || errors.As(err, &pgErr) && pgErr.Field('C') == invalidText {
		return fmt.Errorf("%w: %q", ErrNotFound, id)
	}
	return err
}
