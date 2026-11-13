package listbrokenlinks

import (
	"context"
	"slices"
	"strings"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/notetitle"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readgraph"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/wikilink"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

// The kinds of broken link.
const (
	UNRESOLVED = "unresolved" // [[target]] that no other note has as title or path, a heading the note lacks
	AMBIGUOUS  = "ambiguous"  // [[target]] that several other notes have
	MISSING    = "missing"    // a link to a note trashed or purged since
	FILE       = "file"       // an image or attachment /api/files/<id> the account does not keep: imported from another, purged
)

// Link is a broken link of a live note: its target as written, the label of a link to a missing note, or the address of a missing file.
type Link struct {
	NoteID    string `json:"note_id"`
	NoteTitle string `json:"note_title"`
	NotePath  string `json:"note_path"`
	Target    string `json:"target"`
	Kind      string `json:"kind" enums:"unresolved,ambiguous,missing,file"`
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
	files    map[string]bool
	result   []Link
}

// Process lists the links of the live notes that point to no note, or wait for a title several notes have, and the files they cite that the account lacks.
func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

func (p *run) Run() {
	p.loadNotes()
	p.loadContents()
	p.loadFiles()
	p.check()
}

// Getters & Setters

func (p *run) GetResult() []Link {
	return p.cache.result
}

// Methods

// loadNotes reads the live notes, their titles and paths, sorted by path: the graph's nodes.
func (p *run) loadNotes() {
	graph := readgraph.Process(p.ctx, p.db, false)
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

// loadFiles reads the ids of the files the account keeps.
func (p *run) loadFiles() {
	if p.HasErr() {
		return
	}

	var ids []string
	p.SetErr(p.db.NewSelect().Model((*domain.File)(nil)).Column("id").
		Where("user_id = ?", domain.UserID(p.ctx)).Scan(p.ctx, &ids))
	p.cache.files = make(map[string]bool, len(ids))
	for _, id := range ids {
		p.cache.files[id] = true
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
			name, heading, _ := strings.Cut(target, "#")
			if name = strings.TrimSpace(name); name == "" {
				// [[#Heading]]: a heading of the note itself.
				if !hasHeading(content, heading) {
					add(target, UNRESOLVED)
				}
				continue
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
		for _, link := range wikilink.HeadingLinks(content) {
			if live[link.ID] && !hasHeading(p.cache.contents[link.ID], link.Heading) {
				add(link.Label, UNRESOLVED)
			}
		}
		labels := wikilink.Labels(content)
		for _, id := range wikilink.IDs(content) {
			if !live[strings.ToLower(id)] {
				add(labels[strings.ToLower(id)], MISSING)
			}
		}
		for _, match := range domain.FilePattern.FindAllStringSubmatch(content, -1) {
			if !p.cache.files[match[1]] {
				add(match[0], FILE)
			}
		}
	}
}

// Helpers

// hasHeading tells whether content has heading, as the app compares them: any case, no parentheses.
func hasHeading(content, heading string) bool {
	plain := strings.NewReplacer("(", "", ")", "")
	return slices.ContainsFunc(notetitle.Headings(content), func(h string) bool {
		return strings.EqualFold(strings.TrimSpace(plain.Replace(h)), strings.TrimSpace(plain.Replace(heading)))
	})
}
