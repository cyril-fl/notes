package listbrokenlinks

import (
	"context"
	"strings"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readgraph"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/wikilink"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

// The kinds of broken link.
const (
	UNRESOLVED = "unresolved" // [[target]] that no other note has as title or path
	AMBIGUOUS  = "ambiguous"  // [[target]] that several other notes have
	MISSING    = "missing"    // a link to a note trashed or purged since
)

// Link is a broken link of a live note: its target as written, or the label of a link to a missing note.
type Link struct {
	NoteID    string `json:"note_id"`
	NoteTitle string `json:"note_title"`
	NotePath  string `json:"note_path"`
	Target    string `json:"target"`
	Kind      string `json:"kind" enums:"unresolved,ambiguous,missing"`
}

type run struct {
	plumprimitive.Process
	cache

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
}

type cache struct {
	notes    []readgraph.Note
	contents map[string]string
	result   []Link
}

// Process lists the links of the live notes that point to no note, or wait for a title several notes have.
func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

func (p *run) Run() {
	p.loadNotes()
	p.loadContents()
	p.check()
}

// Getters & Setters

func (p *run) GetResult() []Link {
	return p.cache.result
}

// Methods

// loadNotes reads the live notes, their titles and paths, sorted by path: the graph's nodes.
func (p *run) loadNotes() {
	graph := readgraph.Process(p.ctx, p.db)
	graph.Run()
	if graph.HasErr() {
		p.SetErr(graph.GetErr())
		return
	}
	p.cache.notes = graph.GetResult().Notes
}

// TODO: decrypts every live note on each call, like linknotes' resolveWaiting; a cache or a flag on the note if it gets slow (#23).
func (p *run) loadContents() {
	if p.HasErr() {
		return
	}

	var rows []struct {
		ID      string      `bun:"id"`
		Content secret.Text `bun:"content"`
	}
	p.SetErr(p.db.NewRaw(domain.LiveCTE+`
		SELECT mn.item_id AS id, mn.content FROM metadata_notes AS mn JOIN live ON live.id = mn.item_id`, domain.UserID(p.ctx)).
		Scan(p.ctx, &rows))
	p.cache.contents = make(map[string]string, len(rows))
	for _, row := range rows {
		p.cache.contents[row.ID] = string(row.Content)
	}
}

// check counts, for each [[target]], the other notes with that title (any case) or that path, as linknotes.Resolve
// finds them, and looks for the ids of resolved links among the live notes.
func (p *run) check() {
	if p.HasErr() {
		return
	}

	named := map[string][]string{}
	live := map[string]bool{}
	for _, note := range p.cache.notes {
		named[strings.ToLower(string(note.Title))] = append(named[strings.ToLower(string(note.Title))], note.ID)
		named[note.Path] = append(named[note.Path], note.ID)
		live[note.ID] = true
	}

	for _, note := range p.cache.notes {
		content := p.cache.contents[note.ID]
		add := func(target, kind string) {
			p.cache.result = append(p.cache.result, Link{note.ID, string(note.Title), note.Path, target, kind})
		}
		for _, target := range wikilink.Parse(content) {
			name, _, _ := strings.Cut(target, "#")
			if name = strings.TrimSpace(name); name == "" {
				continue // [[#Heading]]: a heading of the note itself
			}
			if !strings.HasPrefix(name, "/") {
				name = strings.ToLower(name)
			}
			others := 0
			for _, id := range named[name] {
				if id != note.ID {
					others++
				}
			}
			if others == 0 {
				add(target, UNRESOLVED)
			} else if others > 1 {
				add(target, AMBIGUOUS)
			}
		}
		labels := wikilink.Labels(content)
		for _, id := range wikilink.IDs(content) {
			if !live[strings.ToLower(id)] {
				add(labels[strings.ToLower(id)], MISSING)
			}
		}
	}
}
