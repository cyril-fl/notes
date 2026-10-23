package createschema

import (
	"context"
	"database/sql"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
)

type run struct {
	plumprimitive.Process
	cache

	ctx context.Context
	db  *bun.DB
}

type cache struct {
	steps []func(tx bun.Tx)
}

func Process(ctx context.Context, db *bun.DB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

func (p *run) Run() {
	p.registerModels()
	p.createTables()
	p.createIndexes()
	p.runInTx()
}

// Methods
func (p *run) registerModels() {
	domain.RegisterModels(p.db)
}

// createTables and createIndexes only queue their work: runInTx executes it in one transaction.
func (p *run) createTables() {
	p.cache.steps = append(p.cache.steps, func(tx bun.Tx) {
		p.execQueries(getTables(tx))
	})
}

func (p *run) createIndexes() {
	p.cache.steps = append(p.cache.steps, func(tx bun.Tx) {
		p.execQueries(getIndexes(tx))
	})
}

// runInTx runs the queued steps: bun commits, or rolls back if a step recorded an error.
func (p *run) runInTx() {
	p.ClearErr()

	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		p.runStep(tx)
		return p.GetErr()
	}))

}

// execQueries stops at the first error: Postgres aborts the transaction, later statements would only add noise.
func (p *run) execQueries[T query[T]](queries []T) {
	for _, q := range queries {
		if p.HasErr() {
			return
		}

		_, err := q.IfNotExists().Exec(p.ctx)
		p.SetErr(err)
	}
}

func (p *run) runStep(tx bun.Tx) {
	for _, step := range p.cache.steps {
		step(tx)
	}
}

// Helpers

// Order matters for foreign keys.
// FKs are declared here rather than via WithForeignKeys: several are composite (id, kind)
// so the DB itself enforces that metadata and parents match the item kind.
// kind / parent_kind below are DB-only columns (not in the structs): their DEFAULT is the only allowed value.
func getTables(tx bun.Tx) []*bun.CreateTableQuery {
	return []*bun.CreateTableQuery{
		tx.NewCreateTable().Model((*domain.Item)(nil)).
			ColumnExpr("CONSTRAINT items_kind_check CHECK (kind IN (?))", bun.List([]domain.Kind{domain.NOTE, domain.FOLDER})).
			ColumnExpr("parent_kind VARCHAR NOT NULL DEFAULT ? CHECK (parent_kind = ?)", domain.FOLDER, domain.FOLDER).
			ColumnExpr("UNIQUE (id, kind)").
			// MATCH SIMPLE: a NULL parent_id (root item) skips the check.
			ForeignKey(`("parent_id", "parent_kind") REFERENCES "items" ("id", "kind") ON DELETE CASCADE`),
		tx.NewCreateTable().Model((*domain.MetadataNote)(nil)).
			ColumnExpr("kind VARCHAR NOT NULL DEFAULT ? CHECK (kind = ?)", domain.NOTE, domain.NOTE).
			ForeignKey(`("item_id", "kind") REFERENCES "items" ("id", "kind") ON DELETE CASCADE`),
		tx.NewCreateTable().Model((*domain.MetadataFolder)(nil)).
			ColumnExpr("kind VARCHAR NOT NULL DEFAULT ? CHECK (kind = ?)", domain.FOLDER, domain.FOLDER).
			ForeignKey(`("item_id", "kind") REFERENCES "items" ("id", "kind") ON DELETE CASCADE`),
		tx.NewCreateTable().Model((*domain.Tag)(nil)),
		tx.NewCreateTable().Model((*domain.NoteTag)(nil)).
			ForeignKey(`("tag_id") REFERENCES "tags" ("id") ON DELETE CASCADE`).
			ForeignKey(`("note_id") REFERENCES "metadata_notes" ("item_id") ON DELETE CASCADE`),
	}
}

// Postgres does not index FK columns; notes_on_tags PK (tag_id, note_id) only covers tag_id lookups.
func getIndexes(tx bun.Tx) []*bun.CreateIndexQuery {
	return []*bun.CreateIndexQuery{
		tx.NewCreateIndex().Model((*domain.Item)(nil)).Index("idx_items_parent_id").Column("parent_id"),
		tx.NewCreateIndex().Model((*domain.NoteTag)(nil)).Index("idx_notes_on_tags_note_id").Column("note_id"),
		tx.NewCreateIndex().Model((*domain.Tag)(nil)).Index("ux_tags_value_lower").Unique().ColumnExpr("lower(value)"),
	}
}

// Interface

// query is satisfied by both *bun.CreateTableQuery and *bun.CreateIndexQuery.
type query[T any] interface {
	IfNotExists() T
	Exec(ctx context.Context, dest ...any) (sql.Result, error)
}
