package readgraph

import (
	"context"

	"main.go/app/shared/services/secret"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
)

// Note is a node of the graph.
type Note struct {
	ID    string      `json:"id"`
	Title secret.Text `json:"title" swaggertype:"string"`
	Path  string      `json:"path"`
}

// Link joins two notes once per pair (Source < Target): Linked when a [[link]] goes from one to the
// other, Tags the tags they share. One of the two at least.
type Link struct {
	Source string   `json:"source"`
	Target string   `json:"target"`
	Tags   []string `bun:",array" json:"tags"`
	Linked bool     `json:"linked"`
}

// Result is every live note and the links between them.
type Result struct {
	Notes []Note `json:"notes"`
	Links []Link `json:"links"`
}

type run struct {
	plumprimitive.Process
	cache

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
}

type cache struct {
	result Result
}

func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

func (p *run) Run() {
	p.loadNotes()
	p.loadLinks()
}

// Getters & Setters

func (p *run) GetResult() Result {
	return p.cache.result
}

// Methods

// loadNotes joins live to keep live notes only.
func (p *run) loadNotes() {
	p.SetErr(p.db.NewRaw(domain.LiveCTE+`
		SELECT i.id, i.title
		FROM items AS i
		JOIN live ON live.id = i.id
		WHERE i.kind = ?
		ORDER BY i.id`, domain.UserID(p.ctx), domain.NOTE).
		Scan(p.ctx, &p.cache.result.Notes))
	if p.HasErr() {
		return
	}

	paths, err := domain.Paths(p.ctx, p.db)
	p.SetErr(err)
	for i := range p.cache.result.Notes {
		p.cache.result.Notes[i].Path = paths[p.cache.result.Notes[i].ID]
	}
	domain.SortFold(p.cache.result.Notes, func(n Note) string { return n.Path }, nil)
}

// loadLinks pairs the notes sharing a tag (the tag links joined with themselves) with the notes joined by
// a [[link]] in either direction, each pair once. The tags come sealed, opened and ordered here.
func (p *run) loadLinks() {
	if p.HasErr() {
		return
	}

	p.SetErr(p.db.NewRaw(domain.LiveCTE+`,
		shared AS (
			SELECT a.note_id AS source, b.note_id AS target, array_agg(t.value) AS tags
			FROM notes_on_tags AS a
			JOIN notes_on_tags AS b ON b.tag_id = a.tag_id AND a.note_id < b.note_id
			JOIN tags AS t ON t.id = a.tag_id
			GROUP BY a.note_id, b.note_id
		),
		linked AS (
			SELECT DISTINCT least(source_id, target_id) AS source, greatest(source_id, target_id) AS target
			FROM notes_on_links
		)
		SELECT coalesce(shared.source, linked.source) AS source, coalesce(shared.target, linked.target) AS target,
			coalesce(shared.tags, '{}') AS tags, linked.source IS NOT NULL AS linked
		FROM shared FULL JOIN linked ON linked.source = shared.source AND linked.target = shared.target
		JOIN live AS source_live ON source_live.id = coalesce(shared.source, linked.source)
		JOIN live AS target_live ON target_live.id = coalesce(shared.target, linked.target)
		ORDER BY 1, 2`, domain.UserID(p.ctx)).
		Scan(p.ctx, &p.cache.result.Links))

	for i := range p.cache.result.Links {
		if p.HasErr() {
			return
		}
		tags, err := domain.OpenTags(p.cache.result.Links[i].Tags)
		p.cache.result.Links[i].Tags = tags
		p.SetErr(err)
	}
}
