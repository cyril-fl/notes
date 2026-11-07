package editnote

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/hashtag"
	"main.go/app/shared/services/notetitle"
	"main.go/app/shared/services/process/createtag"
	"main.go/app/shared/services/process/linknotes"
)

// invalidText is the Postgres SQLSTATE raised when the id is not a uuid.
const invalidText = "22P02"

var (
	ErrNoValue   = errors.New("no id given")
	ErrNoContent = errors.New("no note content given")
	ErrNotFound  = errors.New("note not found")
)

// Input is the whole new state of the note, its title comes from Content. Content ends up holding
// exactly Tags as #tags, so callers give hashtag.Parse(Content) when only the content changes.
type Input struct {
	ID      string
	Content string
	Tags    []string
}

type run struct {
	plumprimitive.Process
	cache

	ctx   context.Context
	db    bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	input Input
}

type cache struct {
	input Input
	item  domain.Item
	// oldTitle is the title before this edit: the links to the note take the new one.
	oldTitle string
	tags     []domain.Tag
	// unlinked are the tags the note had before this edit.
	unlinked []string
}

func Process(ctx context.Context, db bun.IDB, input Input) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		input: input,
	}
}

func (p *run) Run() {
	p.makeNote()
	p.saveNote()
}

// Getters & Setters

// GetResult returns the updated item.
func (p *run) GetResult() domain.Item {
	return p.cache.item
}

// Methods

func (p *run) makeNote() {
	p.cache.input = Input{
		ID:      strings.TrimSpace(p.input.ID),
		Content: strings.TrimSpace(p.input.Content),
	}
	var tags []string
	for _, tag := range p.input.Tags {
		if tag = strings.TrimSpace(tag); tag == "" {
			continue
		}
		if !hashtag.Valid(tag) {
			p.SetErr(fmt.Errorf("%w: %q", hashtag.ErrInvalid, tag))
			return
		}
		tags = append(tags, tag)
	}

	p.cache.input.Content = hashtag.Sync(p.cache.input.Content, tags)
	p.cache.input.Tags = hashtag.Parse(p.cache.input.Content)

	switch {
	case p.cache.input.ID == "":
		p.SetErr(ErrNoValue)
	case p.cache.input.Content == "":
		p.SetErr(ErrNoContent)
	}
}

// saveNote updates item, metadata, tag and note links in one transaction: a failure changes nothing.
func (p *run) saveNote() {
	if p.HasErr() {
		return
	}

	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, step := range []func(context.Context, bun.Tx) error{p.readOldTitle, p.updateItem, p.updateMetadata, p.replaceTags, p.dropUnusedTags, p.linkNotes} {
			if err := step(ctx, tx); err != nil {
				return err
			}
		}
		return nil
	}))
}

// readOldTitle finds nothing for an unknown id: updateItem then reports it.
func (p *run) readOldTitle(ctx context.Context, tx bun.Tx) error {
	err := tx.NewRaw("SELECT coalesce(title, '') FROM items WHERE id::text = ? AND kind = ? AND user_id = ?", p.cache.input.ID, domain.NOTE, domain.UserID(p.ctx)).Scan(ctx, &p.cache.oldTitle)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}

// updateItem matches a note not in the trash (soft_delete adds deleted_at IS NULL): anything else is not found.
func (p *run) updateItem(ctx context.Context, tx bun.Tx) error {
	err := tx.NewUpdate().
		Model(&p.cache.item).
		Set("title = NULLIF(?, '')", notetitle.From(p.cache.input.Content)). // NULL = untitled, like createnote's nullzero
		Set("updated_at = current_timestamp").
		Where("id = ? AND kind = ? AND user_id = ?", p.cache.input.ID, domain.NOTE, domain.UserID(p.ctx)).
		Returning("id, parent_id, created_at, updated_at, kind, title").
		Scan(ctx)
	return toDomainErr(err, p.cache.input.ID)
}

func (p *run) updateMetadata(ctx context.Context, tx bun.Tx) error {
	_, err := tx.NewUpdate().
		Model((*domain.MetadataNote)(nil)).
		Set("content = ?", p.cache.input.Content).
		Where("item_id = ?", p.cache.item.ID).
		Exec(ctx)
	return err
}

// replaceTags drops every link, creates the unknown tags, then links the given ones.
func (p *run) replaceTags(ctx context.Context, tx bun.Tx) error {
	err := tx.NewDelete().Model((*domain.NoteTag)(nil)).Where("note_id = ?", p.cache.item.ID).Returning("tag_id").Scan(ctx, &p.cache.unlinked)
	if err != nil || len(p.cache.input.Tags) == 0 {
		return err
	}

	create := createtag.Process(ctx, tx, p.cache.input.Tags)
	create.Run()
	if create.HasErr() {
		return create.GetErr()
	}

	// createtag only returns new tags: select again to get the ids of existing ones too.
	err = tx.NewSelect().
		Model(&p.cache.tags).
		Where("user_id = ? AND lower(value) IN (SELECT lower(v) FROM unnest(?::text[]) AS v)", domain.UserID(p.ctx), pgdialect.Array(p.cache.input.Tags)).
		Scan(ctx)
	if err != nil {
		return err
	}

	links := make([]domain.NoteTag, len(p.cache.tags))
	for i, tag := range p.cache.tags {
		links[i] = domain.NoteTag{TagID: tag.ID, NoteID: p.cache.item.ID}
	}
	_, err = tx.NewInsert().Model(&links).Exec(ctx)
	return err
}

// dropUnusedTags deletes the tags this edit left without any note, or each pause while typing #golang would
// leave #g and #go behind. A tag created on its own stays: it was never linked here.
func (p *run) dropUnusedTags(ctx context.Context, tx bun.Tx) error {
	return DropUnusedTags(ctx, tx, p.cache.unlinked)
}

// DropUnusedTags deletes, among ids, the tags left without any note. Also used by the purges of the trash.
func DropUnusedTags(ctx context.Context, db bun.IDB, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	_, err := db.NewDelete().
		Model((*domain.Tag)(nil)).
		Where("t.id IN (?)", bun.In(ids)).
		Where("NOT EXISTS (SELECT 1 FROM notes_on_tags AS nt WHERE nt.tag_id = t.id)").
		Exec(ctx)
	return err
}

// linkNotes resolves the [[links]] of the note, links it, and gives the links to it its new title.
func (p *run) linkNotes(ctx context.Context, tx bun.Tx) error {
	input := linknotes.Input{ID: p.cache.item.ID, Content: p.cache.input.Content, Title: p.cache.item.Title}
	if p.cache.oldTitle != input.Title {
		input.OldTitle = p.cache.oldTitle
	}
	link := linknotes.Process(ctx, tx, input)
	link.Run()
	return link.GetErr()
}

// Helpers

// toDomainErr turns no row, or an id that is not a uuid, into ErrNotFound.
func toDomainErr(err error, id string) error {
	var pgErr pgdriver.Error
	if errors.Is(err, sql.ErrNoRows) || errors.As(err, &pgErr) && pgErr.Field('C') == invalidText {
		return fmt.Errorf("%w: %q", ErrNotFound, id)
	}
	return err
}
