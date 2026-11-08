package readneighbors

import (
	"context"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

// Neighbor is a note sharing at least one tag with the center, with the shared tags.
type Neighbor struct {
	ID     string      `json:"id"`
	Title  secret.Text `json:"title" swaggertype:"string"`
	Path   string      `json:"path"`
	Tags   []string    `bun:",array" json:"tags"`
	Shared int         `json:"shared"`
}

// Linked is a note at the other end of a [[link]].
type Linked struct {
	ID    string      `json:"id"`
	Title secret.Text `json:"title" swaggertype:"string"`
	Path  string      `json:"path"`
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
	paths  map[string]string // read once for the neighbors and the links
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
	p.loadPaths()
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

// loadPaths reads every item's path; the center's already came from readitem.
func (p *run) loadPaths() {
	if p.HasErr() {
		return
	}

	paths, err := domain.Paths(p.ctx, p.db)
	p.cache.paths = paths
	p.SetErr(err)
}

// loadNeighbors joins the tag links with themselves: same tag, other note. Joining live keeps live notes only.
// The tags come sealed, opened and ordered here.
func (p *run) loadNeighbors() {
	if p.HasErr() {
		return
	}

	p.SetErr(p.db.NewRaw(domain.LiveCTE+`
		SELECT i.id, i.title, array_agg(t.value) AS tags, count(*) AS shared
		FROM notes_on_tags AS center
		JOIN notes_on_tags AS other ON other.tag_id = center.tag_id AND other.note_id <> center.note_id
		JOIN tags AS t ON t.id = center.tag_id
		JOIN items AS i ON i.id = other.note_id
		JOIN live ON live.id = i.id
		WHERE center.note_id = ?
		GROUP BY i.id, i.title`, domain.UserID(p.ctx), p.cache.result.Note.Item.ID).
		Scan(p.ctx, &p.cache.result.Neighbors))
	if p.HasErr() {
		return
	}

	for i := range p.cache.result.Neighbors {
		neighbor := &p.cache.result.Neighbors[i]
		neighbor.Path = p.cache.paths[neighbor.ID]
		if !p.HasErr() {
			var err error
			neighbor.Tags, err = domain.OpenTags(neighbor.Tags)
			p.SetErr(err)
		}
	}
	domain.SortFold(p.cache.result.Neighbors, func(n Neighbor) string { return n.Path }, func(a, b Neighbor) int { return b.Shared - a.Shared })
}

// loadLinks reads both ends of the [[links]]; joining live keeps live notes only. Empty, not null, in JSON.
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
		err := p.db.NewRaw(domain.LiveCTE+`
			SELECT i.id, i.title
			FROM notes_on_links AS nl
			JOIN items AS i ON i.id = nl.`+side.to+`
			JOIN live ON live.id = i.id
			WHERE nl.`+side.from+` = ?`, domain.UserID(p.ctx), id).
			Scan(p.ctx, side.dest)
		if err != nil {
			p.SetErr(err)
			return
		}
	}

	for _, linked := range [][]Linked{p.cache.result.Links, p.cache.result.Backlinks} {
		for i := range linked {
			linked[i].Path = p.cache.paths[linked[i].ID]
		}
		domain.SortFold(linked, func(l Linked) string { return l.Path }, nil)
	}
}
