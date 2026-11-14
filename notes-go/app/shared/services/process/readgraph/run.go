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

// tagRow is a tag with its notes, live or not.
type tagRow struct {
	Value   string
	NoteIDs []string `bun:"note_ids,array"`
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
	paths  map[string]string // the live items: a note missing from it is trashed, or in a trashed folder
}

// Process reads the notes sorted by path, or by id with their tags and [[links]] when links is set.
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

// loadNotes keeps the live notes by their paths, cached, not by LiveCTE: walking the folders costs 150 ms at 50,000 notes.
func (p *run) loadNotes() {
	paths, err := domain.Paths(p.ctx, p.db)
	p.SetErr(err)
	if p.HasErr() {
		return
	}
	p.cache.paths = paths

	var notes []Note
	p.SetErr(p.db.NewRaw(`
		SELECT id, title
		FROM items
		WHERE user_id = ? AND kind = ? AND deleted_at IS NULL
		ORDER BY id`, domain.UserID(p.ctx), domain.NOTE).
		Scan(p.ctx, &notes))
	for _, note := range notes {
		if path, ok := paths[note.ID]; ok {
			note.Path = path
			p.cache.result.Notes = append(p.cache.result.Notes, note)
		}
	}
	// With the links, the graph has no order to show: by id, sorting 50,000 paths costs 55 ms.
	if !p.links {
		domain.SortFold(p.cache.result.Notes, func(n Note) string { return n.Path }, nil)
	}
}

// loadTags gives each live note its tags. A tag is one row with its notes: its sealed value read and opened once.
func (p *run) loadTags() {
	if p.HasErr() || !p.links {
		return
	}

	var rows []tagRow
	p.SetErr(p.db.NewRaw(`
		SELECT t.value, array_agg(nt.note_id) AS note_ids
		FROM tags AS t
		JOIN notes_on_tags AS nt ON nt.tag_id = t.id
		WHERE t.user_id = ?
		GROUP BY t.id`, domain.UserID(p.ctx)).
		Scan(p.ctx, &rows))
	if p.HasErr() {
		return
	}

	for i := range rows {
		var err error
		if rows[i].Value, err = secret.Open(rows[i].Value); err != nil {
			p.SetErr(err)
			return
		}
	}
	// The tags sorted once: each note gets its own already in order.
	domain.SortFold(rows, func(row tagRow) string { return row.Value }, nil)

	tags := map[string][]string{}
	for _, row := range rows {
		for _, id := range row.NoteIDs {
			if p.live(id) {
				tags[id] = append(tags[id], row.Value)
			}
		}
	}
	for i := range p.cache.result.Notes {
		p.cache.result.Notes[i].Tags = tags[p.cache.result.Notes[i].ID]
	}
}

// loadLinks reads the [[links]] between live notes, each pair once whichever way it goes.
func (p *run) loadLinks() {
	if p.HasErr() || !p.links {
		return
	}

	var links []Link
	p.SetErr(p.db.NewRaw(`
		SELECT DISTINCT least(nl.source_id, nl.target_id) AS source, greatest(nl.source_id, nl.target_id) AS target
		FROM notes_on_links AS nl
		JOIN items AS i ON i.id = nl.source_id
		WHERE i.user_id = ?
		ORDER BY 1, 2`, domain.UserID(p.ctx)).
		Scan(p.ctx, &links))
	for _, link := range links {
		if p.live(link.Source) && p.live(link.Target) {
			p.cache.result.Links = append(p.cache.result.Links, link)
		}
	}
}

// Helpers

func (p *run) live(id string) bool {
	_, ok := p.cache.paths[id]
	return ok
}
