package linknotes

import (
	"context"
	"slices"
	"strings"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/wikilink"
)

// Input is a note as just saved: its id, content and title ("" = untitled), and the title it had
// before this save when it changed: the links naming either may now resolve differently.
type Input struct {
	ID       string
	Content  string
	Title    string
	OldTitle string
}

type run struct {
	plumprimitive.Process
	cache

	ctx   context.Context
	db    bun.IDB // the caller's transaction: links change with the note, or not at all
	input Input
}

type cache struct {
	before  []string
	after   []string
	changed bool
}

// Process makes the links of a note exactly the notes its [[links]] name, then links again the notes
// naming its title or its old one: written before it existed, or now ambiguous, or no longer.
// shortcut: trashing or restoring a note does not relink the notes naming it; do it there if it matters.
func Process(ctx context.Context, db bun.IDB, input Input) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		input: input,
	}
}

func (p *run) Run() {
	p.replaceLinks()
	p.linkNamers()
}

// Getters & Setters

// GetResult tells whether the links of the note changed: false on a second run.
func (p *run) GetResult() bool {
	return p.cache.changed
}

// Methods

// replaceLinks drops every link of the note, then links the targets that name exactly one live note.
func (p *run) replaceLinks() {
	err := p.db.NewDelete().Model((*domain.NoteLink)(nil)).Where("source_id = ?", p.input.ID).Returning("target_id").Scan(p.ctx, &p.cache.before)
	if err != nil {
		p.SetErr(err)
		return
	}

	for _, target := range wikilink.Parse(p.input.Content) {
		id, err := Resolve(p.ctx, p.db, target)
		if err != nil {
			p.SetErr(err)
			return
		}
		if id != "" && id != p.input.ID && !slices.Contains(p.cache.after, id) {
			p.cache.after = append(p.cache.after, id)
		}
	}

	slices.Sort(p.cache.before)
	slices.Sort(p.cache.after)
	p.cache.changed = !slices.Equal(p.cache.before, p.cache.after)
	if len(p.cache.after) == 0 {
		return
	}

	links := make([]domain.NoteLink, len(p.cache.after))
	for i, target := range p.cache.after {
		links[i] = domain.NoteLink{SourceID: p.input.ID, TargetID: target}
	}
	_, err = p.db.NewInsert().Model(&links).Exec(p.ctx)
	p.SetErr(err)
}

// linkNamers reads only the notes whose text holds one of the titles: resolving every note would be one
// query per link. Each namer gets all its links again, so a title now shared drops the link it resolved to.
func (p *run) linkNamers() {
	for _, title := range []string{p.input.Title, p.input.OldTitle} {
		if p.HasErr() || title == "" {
			continue
		}

		var namers []Input
		err := p.db.NewRaw(`
			SELECT mn.item_id AS id, mn.content FROM metadata_notes AS mn
			WHERE mn.item_id <> ? AND strpos(lower(mn.content), lower(?)) > 0`,
			p.input.ID, title+"]]").
			Scan(p.ctx, &namers)
		if err != nil {
			p.SetErr(err)
			return
		}

		for _, namer := range namers {
			if !slices.ContainsFunc(wikilink.Parse(namer.Content), func(target string) bool { return names(target, title) }) {
				continue
			}
			relink := Process(p.ctx, p.db, Input{ID: namer.ID, Content: namer.Content}) // No title: no namers of namers.
			relink.Run()
			if relink.HasErr() {
				p.SetErr(relink.GetErr())
				return
			}
		}
	}
}

// Helpers

// Resolve gives the id of the one live note a target names, by path when it starts with /, else by
// title whatever its case; "" when no note, or several, match: the link stays unresolved.
func Resolve(ctx context.Context, db bun.IDB, target string) (string, error) {
	var ids []string
	query := domain.PathsCTE + `
		SELECT i.id FROM items AS i JOIN paths ON paths.id = i.id
		WHERE i.kind = ? AND lower(i.title) = lower(?) LIMIT 2`
	if strings.HasPrefix(target, "/") {
		query = domain.PathsCTE + `
		SELECT i.id FROM items AS i JOIN paths ON paths.id = i.id
		WHERE i.kind = ? AND paths.path = ? LIMIT 2`
	}
	if err := db.NewRaw(query, domain.NOTE, target).Scan(ctx, &ids); err != nil {
		return "", err
	}
	if len(ids) != 1 {
		return "", nil
	}
	return ids[0], nil
}

// names tells whether target is title, or a path ending with it.
func names(target, title string) bool {
	return strings.EqualFold(target, title) || strings.HasPrefix(target, "/") && strings.HasSuffix(strings.ToLower(target), "/"+strings.ToLower(title))
}
