package readgraph

import (
	"context"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/process/readitem"
)

// Note is a node of the graph.
type Note struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

// Link joins two notes sharing at least one tag, once per pair (Source < Target), with the shared tags.
type Link struct {
	Source string   `json:"source"`
	Target string   `json:"target"`
	Tags   []string `bun:",array" json:"tags"`
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

// loadNotes joins the paths to keep live notes only.
func (p *run) loadNotes() {
	p.SetErr(p.db.NewRaw(readitem.PathsCTE+`
		SELECT i.id, coalesce(i.title, '') AS title, paths.path
		FROM items AS i
		JOIN paths ON paths.id = i.id
		WHERE i.kind = ?
		ORDER BY lower(paths.path), i.id`, domain.NOTE).
		Scan(p.ctx, &p.cache.result.Notes))
}

// loadLinks joins the tag links with themselves: same tag, other note, each pair once.
func (p *run) loadLinks() {
	if p.HasErr() {
		return
	}

	p.SetErr(p.db.NewRaw(readitem.PathsCTE+`
		SELECT a.note_id AS source, b.note_id AS target,
			array_agg(t.value ORDER BY lower(t.value)) AS tags
		FROM notes_on_tags AS a
		JOIN notes_on_tags AS b ON b.tag_id = a.tag_id AND a.note_id < b.note_id
		JOIN tags AS t ON t.id = a.tag_id
		JOIN paths AS source_path ON source_path.id = a.note_id
		JOIN paths AS target_path ON target_path.id = b.note_id
		GROUP BY a.note_id, b.note_id
		ORDER BY a.note_id, b.note_id`).
		Scan(p.ctx, &p.cache.result.Links))
}
