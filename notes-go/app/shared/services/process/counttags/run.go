package counttags

import (
	"context"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

// TagCount is a tag with the number of live notes carrying it.
type TagCount struct {
	Value string
	Notes int
}

type run struct {
	plumprimitive.Process
	cache

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
}

type cache struct {
	result []TagCount
}

func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

func (p *run) Run() {
	p.countTags()
}

// Getters & Setters

// GetResult returns every tag, most used first; an unused tag counts 0.
func (p *run) GetResult() []TagCount {
	return p.cache.result
}

// Methods

// countTags uses LEFT JOINs so unused tags stay, and counts only notes not in the trash.
func (p *run) countTags() {
	p.SetErr(p.db.NewRaw(`
		SELECT t.value, count(i.id) AS notes
		FROM tags AS t
		LEFT JOIN notes_on_tags AS nt ON nt.tag_id = t.id
		LEFT JOIN items AS i ON i.id = nt.note_id AND i.deleted_at IS NULL
		GROUP BY t.id, t.value
		ORDER BY notes DESC, lower(t.value)`).
		Scan(p.ctx, &p.cache.result))
}
