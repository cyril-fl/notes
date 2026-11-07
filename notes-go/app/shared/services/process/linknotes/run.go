package linknotes

import (
	"context"
	"maps"
	"slices"
	"strings"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/notetitle"
	"main.go/app/shared/services/secret"
	"main.go/app/shared/services/wikilink"
)

// Input is a note as just saved: its id, content and title ("" = untitled), and the title it had
// before this save when it changed.
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
	content string
	before  []string
	after   []string
	changed bool
}

// Process makes the [[links]] of a note that name exactly one note point to its id, links the note to the
// notes its text points to, gives the links to it its new title, and resolves the [[links]] of the other
// notes that were waiting for its title. It also keeps the note's notes_on_files.
func Process(ctx context.Context, db bun.IDB, input Input) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		input: input,
	}
}

func (p *run) Run() {
	p.resolveLinks()
	p.replaceLinks()
	p.replaceFiles()
	p.relabelBacklinks()
	p.resolveWaiting()
}

// Getters & Setters

// GetResult gives the content as saved, its [[links]] resolved, and whether it or its links changed:
// false on a second run.
func (p *run) GetResult() (string, bool) {
	return p.cache.content, p.cache.changed
}

// Methods

// resolveLinks saves the content again when a [[link]] became a link to an id: the note's title follows,
// it may come from a heading holding the link.
func (p *run) resolveLinks() {
	var err error
	p.cache.content = wikilink.Resolve(p.input.Content, func(target string) (string, string, bool) {
		if err != nil {
			return "", "", false
		}
		var id, title string
		id, title, err = Resolve(p.ctx, p.db, target)
		return id, title, err == nil && id != "" && id != p.input.ID
	})
	if err != nil || p.cache.content == p.input.Content {
		p.SetErr(err)
		return
	}

	p.cache.changed = true
	p.SetErr(save(p.ctx, p.db, p.input.ID, p.cache.content))
}

// replaceLinks drops every link of the note, then links the notes its text points to that still exist.
func (p *run) replaceLinks() {
	if p.HasErr() {
		return
	}

	err := p.db.NewDelete().Model((*domain.NoteLink)(nil)).Where("source_id = ?", p.input.ID).Returning("target_id").Scan(p.ctx, &p.cache.before)
	if err != nil {
		p.SetErr(err)
		return
	}

	// A purged note is gone from metadata_notes: its link stays in the text, with no row.
	if ids := wikilink.IDs(p.cache.content); len(ids) > 0 {
		err = p.db.NewRaw(`
			SELECT mn.item_id::text FROM metadata_notes AS mn JOIN items AS i ON i.id = mn.item_id
			WHERE mn.item_id::text IN (?) AND mn.item_id <> ? AND i.user_id = ? ORDER BY 1`, bun.In(ids), p.input.ID, domain.UserID(p.ctx)).
			Scan(p.ctx, &p.cache.after)
		if err != nil {
			p.SetErr(err)
			return
		}
	}

	p.cache.changed = p.cache.changed || !sameSet(p.cache.before, p.cache.after)
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

func (p *run) replaceFiles() {
	if p.HasErr() {
		return
	}

	p.SetErr(Files(p.ctx, p.db, p.input.ID, p.cache.content))
}

// relabelBacklinks writes the new title in the links to the note: they still point to it by id,
// the label is what the CLI and an export show.
func (p *run) relabelBacklinks() {
	if p.HasErr() || p.input.Title == "" || p.input.Title == p.input.OldTitle {
		return
	}

	sources, err := loadNotes(p.ctx, p.db, `
		SELECT mn.item_id AS id, mn.content FROM notes_on_links AS nl
		JOIN metadata_notes AS mn ON mn.item_id = nl.source_id
		WHERE nl.target_id = ?`, p.input.ID)
	if err != nil {
		p.SetErr(err)
		return
	}

	for _, source := range sources {
		if content := wikilink.Relabel(source.Content, p.input.ID, p.input.Title); content != source.Content {
			if err := save(p.ctx, p.db, source.ID, content); err != nil {
				p.SetErr(err)
				return
			}
		}
	}
}

// resolveWaiting resolves only the notes whose text holds [[title or [[/…/title: resolving every note would
// be one query per link. Its old title too: shared before, it may now name one note only.
// shortcut: decrypts every note of the account on each title change, fine for a few thousand notes.
func (p *run) resolveWaiting() {
	titles := []string{}
	for _, title := range []string{p.input.Title, p.input.OldTitle} {
		if title != "" {
			titles = append(titles, strings.ToLower(title)+"]]")
		}
	}
	if p.HasErr() || len(titles) == 0 {
		return
	}

	notes, err := loadNotes(p.ctx, p.db, `
		SELECT mn.item_id AS id, mn.content FROM metadata_notes AS mn JOIN items AS i ON i.id = mn.item_id
		WHERE mn.item_id <> ? AND i.user_id = ?`, p.input.ID, domain.UserID(p.ctx))
	if err != nil {
		p.SetErr(err)
		return
	}

	for _, title := range titles {
		for _, note := range notes {
			if !strings.Contains(strings.ToLower(note.Content), title) {
				continue
			}

			// No title: resolving a note must not resolve the notes waiting for it in turn.
			resolve := Process(p.ctx, p.db, Input{ID: note.ID, Content: note.Content})
			resolve.Run()
			if resolve.HasErr() {
				p.SetErr(resolve.GetErr())
				return
			}
		}
	}
}

// Helpers

// Resolve gives the id and title of the one live note a target names, by path when it starts with /,
// else by title whatever its case; "" when no note, or several, match: the link waits.
func Resolve(ctx context.Context, db bun.IDB, target string) (string, string, error) {
	var notes []struct {
		ID    string `bun:"id"`
		Title string `bun:"title"`
	}
	query := domain.PathsCTE + `
		SELECT i.id, i.title FROM items AS i JOIN paths ON paths.id = i.id
		WHERE i.kind = ? AND lower(i.title) = lower(?) LIMIT 2`
	if strings.HasPrefix(target, "/") {
		query = domain.PathsCTE + `
		SELECT i.id, i.title FROM items AS i JOIN paths ON paths.id = i.id
		WHERE i.kind = ? AND paths.path = ? LIMIT 2`
	}
	if err := db.NewRaw(query, domain.UserID(ctx), domain.NOTE, target).Scan(ctx, &notes); err != nil || len(notes) != 1 {
		return "", "", err
	}
	return notes[0].ID, notes[0].Title, nil
}

// Items runs Process on the notes among items that a trash or a restore just changed: the [[links]] waiting
// for their title may now name one note only.
func Items(ctx context.Context, db bun.IDB, items []domain.Item) error {
	titles := map[string]string{}
	for _, item := range items {
		if item.Kind == domain.NOTE {
			titles[item.ID] = item.Title
		}
	}
	if len(titles) == 0 {
		return nil
	}

	notes, err := loadNotes(ctx, db, "SELECT item_id AS id, content FROM metadata_notes WHERE item_id IN (?)", bun.In(slices.Collect(maps.Keys(titles))))
	if err != nil {
		return err
	}
	for _, note := range notes {
		note.Title = titles[note.ID]
		link := Process(ctx, db, note)
		link.Run()
		if link.HasErr() {
			return link.GetErr()
		}
	}
	return nil
}

// Files links the note id to the images its content cites, in place of those it cited.
// Those of its account only: a note may cite the address of another account's.
func Files(ctx context.Context, db bun.IDB, id, content string) error {
	if _, err := db.NewDelete().Model((*domain.NoteFile)(nil)).Where("note_id = ?", id).Exec(ctx); err != nil {
		return err
	}
	var files []string
	for _, match := range domain.FilePattern.FindAllStringSubmatch(content, -1) {
		files = append(files, match[1])
	}
	if len(files) == 0 {
		return nil
	}
	_, err := db.NewRaw(`
		INSERT INTO notes_on_files (note_id, file_id)
		SELECT DISTINCT i.id, f.id FROM items AS i JOIN files AS f ON f.user_id = i.user_id
		WHERE i.id = ? AND f.id::text IN (?)`, id, bun.In(files)).Exec(ctx)
	return err
}

// loadNotes runs a query giving id and content, and opens the content.
func loadNotes(ctx context.Context, db bun.IDB, query string, args ...any) ([]Input, error) {
	var rows []struct {
		ID      string      `bun:"id"`
		Content secret.Text `bun:"content"`
	}
	if err := db.NewRaw(query, args...).Scan(ctx, &rows); err != nil {
		return nil, err
	}
	notes := make([]Input, len(rows))
	for i, row := range rows {
		notes[i] = Input{ID: row.ID, Content: string(row.Content)}
	}
	return notes, nil
}

// save writes content and the title it gives; updated_at stays, the user did not write this change.
// WhereAllWithDeleted: a trashed note keeps its links right too.
func save(ctx context.Context, db bun.IDB, id, content string) error {
	if _, err := db.NewUpdate().Model((*domain.MetadataNote)(nil)).Set("content = ?", secret.Text(content)).Where("item_id = ?", id).Exec(ctx); err != nil {
		return err
	}
	_, err := db.NewUpdate().Model((*domain.Item)(nil)).Set("title = NULLIF(?, '')", notetitle.From(content)).Where("id = ?", id).WhereAllWithDeleted().Exec(ctx)
	return err
}

// sameSet compares whatever the order.
func sameSet(a, b []string) bool {
	a, b = slices.Clone(a), slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}
