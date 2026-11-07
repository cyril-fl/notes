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
	ID     string   `json:"id"`
	Title  string   `json:"title"`
	Path   string   `json:"path"`
	Tags   []string `bun:",array" json:"tags"`
	Shared int      `json:"shared"`
}

// Linked is a note at the other end of a [[link]].
type Linked struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

// Result is the center note (path, content, tags), its neighbors (most shared tags first),
// the notes its [[links]] name, and the notes naming it.
type Result struct {
	Note      readitem.Result `json:"note"`
	Neighbors []Neighbor      `json:"neighbors"`
	Links     []Linked        `json:"links"`
	Backlinks []Linked        `json:"backlinks"`
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
	p.loadLinks()
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

	p.SetErr(p.db.NewRaw(domain.PathsCTE+`
		SELECT i.id, coalesce(i.title, '') AS title, paths.path,
			array_agg(t.value ORDER BY lower(t.value)) AS tags, count(*) AS shared
		FROM notes_on_tags AS center
		JOIN notes_on_tags AS other ON other.tag_id = center.tag_id AND other.note_id <> center.note_id
		JOIN tags AS t ON t.id = center.tag_id
		JOIN items AS i ON i.id = other.note_id
		JOIN paths ON paths.id = i.id
		WHERE center.note_id = ?
		GROUP BY i.id, i.title, paths.path
		ORDER BY shared DESC, lower(paths.path)`, domain.UserID(p.ctx), p.cache.result.Note.Item.ID).
		Scan(p.ctx, &p.cache.result.Neighbors))
}

// loadLinks reads both ends of the [[links]]; joining paths keeps live notes only. Empty, not null, in JSON.
func (p *run) loadLinks() {
	if p.HasErr() {
		return
	}

	id := p.cache.result.Note.Item.ID
	p.cache.result.Links, p.cache.result.Backlinks = []Linked{}, []Linked{}
	for _, side := range []struct {
		from, to string
		dest     *[]Linked
	}{{"source_id", "target_id", &p.cache.result.Links}, {"target_id", "source_id", &p.cache.result.Backlinks}} {
		err := p.db.NewRaw(domain.PathsCTE+`
			SELECT i.id, coalesce(i.title, '') AS title, paths.path
			FROM notes_on_links AS nl
			JOIN items AS i ON i.id = nl.`+side.to+`
			JOIN paths ON paths.id = i.id
			WHERE nl.`+side.from+` = ?
			ORDER BY lower(paths.path)`, domain.UserID(p.ctx), id).
			Scan(p.ctx, side.dest)
		if err != nil {
			p.SetErr(err)
			return
		}
	}
}
