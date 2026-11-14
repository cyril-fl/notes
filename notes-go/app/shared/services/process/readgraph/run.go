package readgraph

import (
	"context"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/forcelayout"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

// Note is a node of the graph, with its tags: the tags are nodes too, joined to their notes.
// Laid out (layoutgraph): its place in % of the view, never 0 (a margin keeps it off the edges), and the pull of each tag.
type Note struct {
	ID           string      `json:"id"`
	Title        secret.Text `json:"title" swaggertype:"string"`
	Path         string      `json:"path"`
	Tags         []string    `json:"tags,omitempty" binding:"optional"`
	X            float64     `json:"x,omitempty" binding:"optional"`
	Y            float64     `json:"y,omitempty" binding:"optional"`
	TagStrengths []float64   `json:"tag_strengths,omitempty" binding:"optional"` // in the order of Tags
}

// Link is a [[link]] between two notes, once per pair (Source < Target) whichever way it goes. Laid out: its pull.
type Link struct {
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Strength float64 `json:"strength,omitempty" binding:"optional"`
}

// Result is every live note, with its tags, and the [[links]] between them. Laid out: the place of each tag, by name.
// Its size follows the tags put on notes, not the pairs of notes sharing one (n² for a tag on n notes).
type Result struct {
	Notes        []Note                       `json:"notes"`
	Links        []Link                       `json:"links"`
	TagPositions map[string]forcelayout.Point `json:"tag_positions,omitempty" binding:"optional"`
}

type run struct {
	plumprimitive.Process
	cache

	ctx   context.Context
	db    bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	links bool
}

type cache struct {
	result Result
}

// Process reads the notes, and their tags and [[links]] when links is set.
func Process(ctx context.Context, db bun.IDB, links bool) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		links: links,
	}
}

func (p *run) Run() {
	p.loadNotes()
	p.loadTags()
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

// loadTags gives each live note its tags. A tag is one row, sealed once: opened once, whatever its number of notes.
func (p *run) loadTags() {
	if p.HasErr() || !p.links {
		return
	}

	var rows []struct {
		NoteID string
		Value  string
	}
	p.SetErr(p.db.NewRaw(domain.LiveCTE+`
		SELECT nt.note_id, t.value
		FROM notes_on_tags AS nt
		JOIN live ON live.id = nt.note_id
		JOIN tags AS t ON t.id = nt.tag_id`, domain.UserID(p.ctx)).
		Scan(p.ctx, &rows))
	if p.HasErr() {
		return
	}

	opened := map[string]string{}
	tags := map[string][]string{}
	for _, row := range rows {
		text, ok := opened[row.Value]
		if !ok {
			var err error
			if text, err = secret.Open(row.Value); err != nil {
				p.SetErr(err)
				return
			}
			opened[row.Value] = text
		}
		tags[row.NoteID] = append(tags[row.NoteID], text)
	}
	for i := range p.cache.result.Notes {
		note := &p.cache.result.Notes[i]
		note.Tags = tags[note.ID]
		domain.SortFold(note.Tags, func(v string) string { return v }, nil)
	}
}

// loadLinks reads the [[links]] between live notes, each pair once whichever way it goes.
func (p *run) loadLinks() {
	if p.HasErr() || !p.links {
		return
	}

	p.SetErr(p.db.NewRaw(domain.LiveCTE+`
		SELECT DISTINCT least(nl.source_id, nl.target_id) AS source, greatest(nl.source_id, nl.target_id) AS target
		FROM notes_on_links AS nl
		JOIN live AS source_live ON source_live.id = nl.source_id
		JOIN live AS target_live ON target_live.id = nl.target_id
		ORDER BY 1, 2`, domain.UserID(p.ctx)).
		Scan(p.ctx, &p.cache.result.Links))
}
