package synctags

import (
	"context"
	"fmt"
	"slices"
	"strings"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/hashtag"
	"main.go/app/shared/services/process/createtag"
)

// note is a note, trashed ones included, with the values of its tags.
type note struct {
	ID      string   `bun:"item_id"`
	Content string   `bun:"content"`
	Tags    []string `bun:"tags,array"`
}

type run struct {
	plumprimitive.Process
	cache

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
}

type cache struct {
	notes   []note
	changed int
}

// Process brings notes made before tags came from the text in line: their tags and #tags become the union of both.
func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

func (p *run) Run() {
	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if err := p.loadNotes(ctx, tx); err != nil {
			return err
		}
		for _, current := range p.cache.notes {
			if err := p.syncNote(ctx, tx, current); err != nil {
				return err
			}
		}
		return nil
	}))
}

// Getters & Setters

// GetResult returns how many notes changed: 0 on a second run.
func (p *run) GetResult() int {
	return p.cache.changed
}

// Methods

func (p *run) loadNotes(ctx context.Context, tx bun.Tx) error {
	return tx.NewRaw(`
		SELECT mn.item_id, mn.content, coalesce(array_agg(t.value) FILTER (WHERE t.id IS NOT NULL), '{}') AS tags
		FROM metadata_notes AS mn
		LEFT JOIN notes_on_tags AS nt ON nt.note_id = mn.item_id
		LEFT JOIN tags AS t ON t.id = nt.tag_id
		WHERE mn.item_id IN (SELECT id FROM items WHERE user_id = ?)
		GROUP BY mn.item_id, mn.content`, domain.UserID(p.ctx)).
		Scan(ctx, &p.cache.notes)
}

// syncNote leaves updated_at alone: nothing the user wrote changed.
func (p *run) syncNote(ctx context.Context, tx bun.Tx, current note) error {
	for _, tag := range current.Tags {
		if !hashtag.Valid(tag) {
			return fmt.Errorf("rename this tag first: %w: %q", hashtag.ErrInvalid, tag)
		}
	}

	content := hashtag.Sync(current.Content, append(hashtag.Parse(current.Content), current.Tags...))
	tags := hashtag.Parse(content)
	if content == current.Content && sameTags(tags, current.Tags) {
		return nil
	}
	p.cache.changed++

	_, err := tx.NewUpdate().
		Model((*domain.MetadataNote)(nil)).
		Set("content = ?", content).
		Where("item_id = ?", current.ID).
		Exec(ctx)
	if err != nil {
		return err
	}
	return linkTags(p.ctx, tx, current.ID, tags)
}

// Helpers

func linkTags(ctx context.Context, tx bun.Tx, id string, values []string) error {
	create := createtag.Process(ctx, tx, values)
	create.Run()
	if create.HasErr() {
		return create.GetErr()
	}

	_, err := tx.NewRaw(`
		INSERT INTO notes_on_tags (tag_id, note_id)
		SELECT t.id, ? FROM tags AS t WHERE t.user_id = ? AND lower(t.value) IN (SELECT lower(v) FROM unnest(?::text[]) AS v)
		ON CONFLICT DO NOTHING`, id, domain.UserID(ctx), pgdialect.Array(values)).
		Exec(ctx)
	return err
}

// sameTags compares whatever the order and the case: a linked tag and its #tag can differ in case.
func sameTags(tags, linked []string) bool {
	return len(tags) == len(linked) && !slices.ContainsFunc(tags, func(tag string) bool {
		return !slices.ContainsFunc(linked, func(value string) bool { return strings.EqualFold(value, tag) })
	})
}
