package createschema

import (
	"context"
	"database/sql"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
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
	p.createTriggers()
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

// createTriggers replaces the triggers each time: CREATE OR REPLACE keeps them up to date without migrations.
func (p *run) createTriggers() {
	p.cache.steps = append(p.cache.steps, func(tx bun.Tx) {
		for _, trigger := range triggers {
			if p.HasErr() {
				return
			}
			_, err := tx.ExecContext(p.ctx, trigger)
			p.SetErr(err)
		}
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
		tx.NewCreateTable().Model((*domain.User)(nil)),
		tx.NewCreateTable().Model((*domain.Session)(nil)).
			ForeignKey(`("user_id") REFERENCES "users" ("id") ON DELETE CASCADE`),
		tx.NewCreateTable().Model((*domain.Item)(nil)).
			ForeignKey(`("user_id") REFERENCES "users" ("id") ON DELETE CASCADE`).
			ColumnExpr("CONSTRAINT items_kind_check CHECK (kind IN (?))", bun.List([]domain.Kind{domain.NOTE, domain.FOLDER})).
			ColumnExpr("parent_kind VARCHAR NOT NULL DEFAULT ? CHECK (parent_kind = ?)", domain.FOLDER, domain.FOLDER).
			ColumnExpr("UNIQUE (id, kind)").
			ColumnExpr("CONSTRAINT items_id_kind_user_key UNIQUE (id, kind, user_id)").
			// The parent is a folder of the same account: one of another account is refused like a missing one.
			// MATCH SIMPLE: a NULL parent_id (root item) skips the check.
			ColumnExpr(`CONSTRAINT items_parent_user_fkey FOREIGN KEY ("parent_id", "parent_kind", "user_id") REFERENCES "items" ("id", "kind", "user_id") ON DELETE CASCADE`),
		tx.NewCreateTable().Model((*domain.MetadataNote)(nil)).
			ColumnExpr("kind VARCHAR NOT NULL DEFAULT ? CHECK (kind = ?)", domain.NOTE, domain.NOTE).
			ForeignKey(`("item_id", "kind") REFERENCES "items" ("id", "kind") ON DELETE CASCADE`),
		tx.NewCreateTable().Model((*domain.MetadataFolder)(nil)).
			ColumnExpr("kind VARCHAR NOT NULL DEFAULT ? CHECK (kind = ?)", domain.FOLDER, domain.FOLDER).
			ForeignKey(`("item_id", "kind") REFERENCES "items" ("id", "kind") ON DELETE CASCADE`),
		tx.NewCreateTable().Model((*domain.Tag)(nil)).
			ForeignKey(`("user_id") REFERENCES "users" ("id") ON DELETE CASCADE`),
		tx.NewCreateTable().Model((*domain.NoteTag)(nil)).
			ForeignKey(`("tag_id") REFERENCES "tags" ("id") ON DELETE CASCADE`).
			ForeignKey(`("note_id") REFERENCES "metadata_notes" ("item_id") ON DELETE CASCADE`),
		tx.NewCreateTable().Model((*domain.NoteLink)(nil)).
			ForeignKey(`("source_id") REFERENCES "metadata_notes" ("item_id") ON DELETE CASCADE`).
			ForeignKey(`("target_id") REFERENCES "metadata_notes" ("item_id") ON DELETE CASCADE`),
		tx.NewCreateTable().Model((*domain.NoteVersion)(nil)).
			ForeignKey(`("item_id") REFERENCES "metadata_notes" ("item_id") ON DELETE CASCADE`),
		// TODO: deleting an account leaves its objects in Silo, remove the prefix <user_id>/ if accounts ever get deleted.
		tx.NewCreateTable().Model((*domain.File)(nil)).
			ForeignKey(`("user_id") REFERENCES "users" ("id") ON DELETE CASCADE`),
		tx.NewCreateTable().Model((*domain.NoteFile)(nil)).
			ForeignKey(`("note_id") REFERENCES "metadata_notes" ("item_id") ON DELETE CASCADE`).
			ForeignKey(`("file_id") REFERENCES "files" ("id") ON DELETE CASCADE`),
	}
}

// Postgres does not index FK columns; notes_on_tags PK (tag_id, note_id) only covers tag_id lookups.
func getIndexes(tx bun.Tx) []*bun.CreateIndexQuery {
	return []*bun.CreateIndexQuery{
		tx.NewCreateIndex().Model((*domain.Item)(nil)).Index("idx_items_parent_id").Column("parent_id"),
		tx.NewCreateIndex().Model((*domain.NoteTag)(nil)).Index("idx_notes_on_tags_note_id").Column("note_id"),
		tx.NewCreateIndex().Model((*domain.User)(nil)).Index("ux_users_name_lower").Unique().ColumnExpr("lower(name)"),
		tx.NewCreateIndex().Model((*domain.Item)(nil)).Index("idx_items_user_id").Column("user_id"),
		// Values and titles are sealed: their blinds, lowered before the HMAC, keep them unique whatever their case.
		tx.NewCreateIndex().Model((*domain.Tag)(nil)).Index("ux_tags_user_value_hash").Unique().ColumnExpr("user_id, value_hash"),
		// A root folder has no parent_id: coalesce gives the root one, so its folders are unique too.
		tx.NewCreateIndex().Model((*domain.Item)(nil)).Index(domain.FolderNameIndex).Unique().
			ColumnExpr("user_id, coalesce(parent_id, '00000000-0000-0000-0000-000000000000'), title_hash").
			Where("kind = ? AND deleted_at IS NULL", domain.FOLDER),
		// The primary key (source_id, target_id) only covers the links of a note, not its backlinks.
		tx.NewCreateIndex().Model((*domain.NoteLink)(nil)).Index("idx_notes_on_links_target_id").Column("target_id"),
		tx.NewCreateIndex().Model((*domain.NoteFile)(nil)).Index("idx_notes_on_files_file_id").Column("file_id"),
		tx.NewCreateIndex().Model((*domain.NoteVersion)(nil)).Index("idx_note_versions_item_id").ColumnExpr("item_id, created_at"),
	}
}

// triggers: an item deleted for good (purge, emptied trash, or a child of either by cascade) hands its images to
// purgefiles at once with orphaned_at = -infinity. BEFORE: the note's notes_on_files and versions are still there; the cascade removes them after.
var triggers = []string{
	`CREATE OR REPLACE FUNCTION orphan_files_of_item() RETURNS trigger AS $$
	BEGIN
		UPDATE files SET orphaned_at = '-infinity'
		WHERE id IN (SELECT file_id FROM notes_on_files WHERE note_id = OLD.id
			UNION SELECT unnest(file_ids) FROM note_versions WHERE item_id = OLD.id);
		RETURN OLD;
	END $$ LANGUAGE plpgsql`,
	`CREATE OR REPLACE TRIGGER items_orphan_files BEFORE DELETE ON items FOR EACH ROW EXECUTE FUNCTION orphan_files_of_item()`,
}

// Interface

// query is satisfied by both *bun.CreateTableQuery and *bun.CreateIndexQuery.
type query[T any] interface {
	IfNotExists() T
	Exec(ctx context.Context, dest ...any) (sql.Result, error)
}
