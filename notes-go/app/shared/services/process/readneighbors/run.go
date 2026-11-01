package readneighbors

import (
	"context"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/process/readitem"
)

// Neighbor is a note sharing at least one tag with the center, with the shared tags.
type Neighbor struct {
	ID     string
	Title  string
	Path   string
	Tags   []string `bun:",array"`
	Shared int
}

// Result is the center note (path, content, tags) and its neighbors, most shared tags first.
type Result struct {
	Note      readitem.Result
	Neighbors []Neighbor
}

type run struct {
	plumprimitive.Process
	cache

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	id  string
}

type cache struct {
	result Result
}

func Process(ctx context.Context, db bun.IDB, id string) *run {
	return &run{
		ctx: ctx,
		db:  db,
		id:  id,
	}
}

func (p *run) Run() {
	p.loadNote()
	p.loadNeighbors()
}

// Getters & Setters

func (p *run) GetResult() Result {
	return p.cache.result
}

// Methods

// loadNote reuses readitem: same errors (ErrNoValue, ErrNotFound) for a missing or trashed note.
func (p *run) loadNote() {
	process := readitem.Process(p.ctx, p.db, domain.NOTE, p.id)
	process.Run()
	p.cache.result.Note = process.GetResult()
	p.SetErr(process.GetErr())
}

// loadNeighbors joins the tag links with themselves: same tag, other note. Joining paths keeps live notes only.
func (p *run) loadNeighbors() {
	if p.HasErr() {
		return
	}

	p.SetErr(p.db.NewRaw(readitem.PathsCTE+`
		SELECT i.id, coalesce(i.title, '') AS title, paths.path,
			array_agg(t.value ORDER BY lower(t.value)) AS tags, count(*) AS shared
		FROM notes_on_tags AS center
		JOIN notes_on_tags AS other ON other.tag_id = center.tag_id AND other.note_id <> center.note_id
		JOIN tags AS t ON t.id = center.tag_id
		JOIN items AS i ON i.id = other.note_id
		JOIN paths ON paths.id = i.id
		WHERE center.note_id = ?
		GROUP BY i.id, i.title, paths.path
		ORDER BY shared DESC, lower(paths.path)`, p.cache.result.Note.Item.ID).
		Scan(p.ctx, &p.cache.result.Neighbors))
}
