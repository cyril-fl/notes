// Package runsavedsearch replays a domain.SavedSearch's criteria: never a stored result, always run again.
package runsavedsearch

import (
	"context"
	"strings"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/searchnotes"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

// Match is a note found by the saved search, with the content line where the text criterion was found.
type Match struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Path     string `json:"path"`
	Snippet  string `json:"snippet,omitempty"`
	Readonly bool   `json:"readonly,omitempty" binding:"optional"`
}

type run struct {
	plumprimitive.Process
	cache

	ctx    context.Context
	db     bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	search domain.SavedSearch
}

type cache struct {
	rows   []row
	result []Match
}

// row is what the SQL criteria give: already narrowed to what tags/dates/folders keep, before any text match.
type row struct {
	ID       string
	Title    secret.Text
	Path     string
	Content  secret.Text
	Readonly bool
}

func Process(ctx context.Context, db bun.IDB, search domain.SavedSearch) *run {
	return &run{
		ctx:    ctx,
		db:     db,
		search: search,
	}
}

func (p *run) Run() {
	p.queryCandidates()
	p.filterText()
}

// Getters & Setters

// GetResult returns the matching notes, sorted by path, then by title match when a text criterion narrowed them.
func (p *run) GetResult() []Match {
	return p.cache.result
}

// Methods

// queryCandidates runs every SQL-filterable criterion (tags, dates, folders): the text criterion, if any, only
// ever narrows this candidate set in Go, after it is decrypted, instead of scanning the whole account.
func (p *run) queryCandidates() {
	query := domain.LiveCTE + `
		SELECT i.id, i.title, mn.content, mn.readonly
		FROM items AS i
		JOIN metadata_notes AS mn ON mn.item_id = i.id
		JOIN live ON live.id = i.id
		WHERE i.kind = ?`
	args := []any{domain.UserID(p.ctx), domain.NOTE}

	if len(p.search.TagIDs) > 0 {
		if p.search.TagMode == "all" {
			query += ` AND i.id IN (
				SELECT note_id FROM notes_on_tags WHERE tag_id IN (?) GROUP BY note_id HAVING COUNT(DISTINCT tag_id) = ?)`
			args = append(args, bun.In(p.search.TagIDs), len(p.search.TagIDs))
		} else {
			query += ` AND i.id IN (SELECT note_id FROM notes_on_tags WHERE tag_id IN (?))`
			args = append(args, bun.In(p.search.TagIDs))
		}
	}
	if p.search.CreatedFrom != nil {
		query += ` AND i.created_at >= ?`
		args = append(args, *p.search.CreatedFrom)
	}
	if p.search.CreatedTo != nil {
		query += ` AND i.created_at <= ?`
		args = append(args, *p.search.CreatedTo)
	}
	if p.search.EditedFrom != nil {
		query += ` AND i.updated_at >= ?`
		args = append(args, *p.search.EditedFrom)
	}
	if p.search.EditedTo != nil {
		query += ` AND i.updated_at <= ?`
		args = append(args, *p.search.EditedTo)
	}
	if len(p.search.FolderIDs) > 0 {
		tree := `(WITH RECURSIVE tree AS (
			SELECT id FROM items WHERE id IN (?) AND user_id = ?
			UNION ALL SELECT child.id FROM items AS child JOIN tree ON child.parent_id = tree.id
		) SELECT id FROM tree)`
		if p.search.FolderMode == "exclude" {
			query += ` AND i.id NOT IN ` + tree
		} else {
			query += ` AND i.id IN ` + tree
		}
		args = append(args, bun.In(p.search.FolderIDs), domain.UserID(p.ctx))
	}

	p.SetErr(p.db.NewRaw(query, args...).Scan(p.ctx, &p.cache.rows))
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

// filterText keeps every candidate when there is no text criterion: the SQL criteria already did the filtering.
func (p *run) filterText() {
	if p.HasErr() {
		return
	}

	text := strings.TrimSpace(p.search.Text)
	if text == "" {
		for _, r := range p.cache.rows {
			p.cache.result = append(p.cache.result, Match{ID: r.ID, Title: string(r.Title), Path: r.Path, Readonly: r.Readonly})
		}
		return
	}

	lower := strings.ToLower(text)
	inTitle := func(r row) bool { return strings.Contains(strings.ToLower(string(r.Title)), lower) }
	for _, r := range p.cache.rows {
		if inTitle(r) || strings.Contains(strings.ToLower(string(r.Content)), lower) {
			p.cache.result = append(p.cache.result, Match{
				ID: r.ID, Title: string(r.Title), Path: r.Path, Readonly: r.Readonly,
				Snippet: searchnotes.Snippet(string(r.Content), text),
			})
		}
	}
}
