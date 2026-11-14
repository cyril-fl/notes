package searchnotes

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

// snippetWidth is how many characters of content a match shows around the text found.
const snippetWidth = 60

var ErrNoValue = errors.New("no text to search")

// Match is a note containing the text, with the content line where it was found.
type Match struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Path     string `json:"path"`
	Snippet  string `json:"snippet"`
	Readonly bool   `json:"readonly,omitempty" binding:"optional"`
}

type run struct {
	plumprimitive.Process
	cache

	ctx  context.Context
	db   bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	text string
}

type cache struct {
	text   string
	rows   []row
	result []Match
}

// row is what the query returns: the whole content, cut down to a snippet in makeMatches.
type row struct {
	ID       string
	Title    secret.Text
	Path     string
	Content  secret.Text
	Readonly bool
}

func Process(ctx context.Context, db bun.IDB, text string) *run {
	return &run{
		ctx:  ctx,
		db:   db,
		text: text,
	}
}

func (p *run) Run() {
	p.makeText()
	p.searchNotes()
	p.makeMatches()
}

// Getters & Setters

// GetResult returns the matching notes: title matches first, then by path.
func (p *run) GetResult() []Match {
	return p.cache.result
}

// Methods

func (p *run) makeText() {
	p.cache.text = strings.TrimSpace(p.text)
	if p.cache.text == "" {
		p.SetErr(ErrNoValue)
	}
}

// searchNotes loads every live note, by path: the content is sealed, makeMatches searches it once opened.
// Joining live keeps live notes only.
// TODO: decrypts every note of the account on each search, 50 ms for 5,000 notes (#145); index the words past ~50,000.
func (p *run) searchNotes() {
	if p.HasErr() {
		return
	}

	p.SetErr(p.db.NewRaw(domain.LiveCTE+`
		SELECT i.id, i.title, mn.content, mn.readonly
		FROM items AS i
		JOIN metadata_notes AS mn ON mn.item_id = i.id
		JOIN live ON live.id = i.id
		WHERE i.kind = ?`,
		domain.UserID(p.ctx), domain.NOTE).
		Scan(p.ctx, &p.cache.rows))
	if p.HasErr() {
		return
	}

	paths, err := domain.Paths(p.ctx, p.db)
	p.SetErr(err)
	for i := range p.cache.rows {
		p.cache.rows[i].Path = paths[p.cache.rows[i].ID]
	}
	domain.SortFold(p.cache.rows, func(r row) string { return r.Path }, nil)
}

// makeMatches keeps the notes holding the text whatever its case, those with it in their title first.
func (p *run) makeMatches() {
	if p.HasErr() {
		return
	}

	text := strings.ToLower(p.cache.text)
	inTitle := func(r row) bool { return strings.Contains(strings.ToLower(string(r.Title)), text) }
	for _, r := range p.cache.rows {
		if inTitle(r) || strings.Contains(strings.ToLower(string(r.Content)), text) {
			p.cache.result = append(p.cache.result, Match{ID: r.ID, Title: string(r.Title), Path: r.Path, Snippet: Snippet(string(r.Content), p.cache.text), Readonly: r.Readonly})
		}
	}
	slices.SortStableFunc(p.cache.result, func(a, b Match) int {
		return boolOrder(strings.Contains(strings.ToLower(b.Title), text)) - boolOrder(strings.Contains(strings.ToLower(a.Title), text))
	})
}

// Helpers

func boolOrder(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Snippet is the first content line containing text, cut around it; the first line when only the title matched.
func Snippet(content, text string) string {
	lines := strings.Split(content, "\n")
	line, at := strings.TrimSpace(lines[0]), 0
	for _, l := range lines {
		if i := strings.Index(strings.ToLower(l), strings.ToLower(text)); i >= 0 {
			line, at = l, i
			break
		}
	}
	return Cut(line, at)
}

// Cut is a content line cut to snippetWidth around the byte at, with … where cut.
func Cut(line string, at int) string {
	runes := []rune(line)
	if len(runes) <= snippetWidth {
		return strings.TrimSpace(line)
	}
	start := max(0, len([]rune(line[:at]))-snippetWidth/3)
	end := min(len(runes), start+snippetWidth)
	cut := strings.TrimSpace(string(runes[start:end]))
	if start > 0 {
		cut = "…" + cut
	}
	if end < len(runes) {
		cut += "…"
	}
	return cut
}
