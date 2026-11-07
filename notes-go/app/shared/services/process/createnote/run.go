package createnote

import (
	"context"
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
	"main.go/app/shared/services/secret"
)

var (
	ErrNoContent      = errors.New("no note content given")
	ErrParentNotFound = errors.New("parent folder not found")
)

// foreignKeyViolation is the Postgres SQLSTATE raised here by items_parent_user_fkey (a missing parent, or another account's).
const foreignKeyViolation = "23503"

// Input is a note as the user gave it: Tags are written into Content as #tags, the title comes from Content.
type Input struct {
	Content  string   `json:"content"`
	ParentID string   `json:"parent_id" binding:"optional"`
	Tags     []string `json:"tags" binding:"optional"`
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
	title string
	item  domain.Item
	tags  []domain.Tag
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

// GetResult returns the created item, with its generated ID.
func (p *run) GetResult() domain.Item {
	return p.cache.item
}

// Methods

func (p *run) makeNote() {
	p.cache.input = Input{
		Content:  strings.TrimSpace(p.input.Content),
		ParentID: strings.TrimSpace(p.input.ParentID),
	}

	if p.cache.input.Content == "" {
		p.SetErr(ErrNoContent)
		return
	}

	tags := hashtag.Parse(p.cache.input.Content)
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
	p.cache.title = notetitle.From(p.cache.input.Content)
}

// saveNote writes item, metadata, tag and note links in one transaction: a failure leaves no half note.
func (p *run) saveNote() {
	if p.HasErr() {
		return
	}

	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, step := range []func(context.Context, bun.Tx) error{p.insertItem, p.insertMetadata, p.linkTags, p.linkNotes} {
			if err := step(ctx, tx); err != nil {
				return err
			}
		}
		return nil
	}))
}

func (p *run) insertItem(ctx context.Context, tx bun.Tx) error {
	p.cache.item = domain.Item{UserID: domain.UserID(p.ctx), Kind: domain.NOTE, Title: p.cache.title}
	if p.cache.input.ParentID != "" {
		p.cache.item.ParentID = &p.cache.input.ParentID
	}

	// Not "*": parent_kind is a DB-only column, Item cannot scan it.
	err := tx.NewInsert().Model(&p.cache.item).Returning("id, created_at").Scan(ctx)
	return toDomainErr(err, p.cache.input.ParentID)
}

func (p *run) insertMetadata(ctx context.Context, tx bun.Tx) error {
	content, err := secret.Seal(domain.UserID(p.ctx), p.cache.input.Content)
	if err != nil {
		return err
	}
	metadata := domain.MetadataNote{
		Metadata: domain.Metadata{ItemID: p.cache.item.ID},
		Content:  content,
	}

	_, err = tx.NewInsert().Model(&metadata).Exec(ctx)
	return err
}

// linkTags creates the unknown tags, then links every given tag to the note.
func (p *run) linkTags(ctx context.Context, tx bun.Tx) error {
	if len(p.cache.input.Tags) == 0 {
		return nil
	}

	create := createtag.Process(ctx, tx, p.cache.input.Tags)
	create.Run()
	if create.HasErr() {
		return create.GetErr()
	}

	// createtag only returns new tags: select again to get the ids of existing ones too.
	err := tx.NewSelect().
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

// linkNotes resolves the [[links]] of the note, links it, and resolves the [[links]] waiting for its title.
func (p *run) linkNotes(ctx context.Context, tx bun.Tx) error {
	link := linknotes.Process(ctx, tx, linknotes.Input{ID: p.cache.item.ID, Content: p.cache.input.Content, Title: p.cache.title})
	link.Run()
	return link.GetErr()
}

// Helpers

// toDomainErr turns the parent FK violation into ErrParentNotFound: a missing id, or one that is not a folder.
func toDomainErr(err error, parentID string) error {
	var pgErr pgdriver.Error

	if errors.As(err, &pgErr) && pgErr.Field('C') == foreignKeyViolation {
		return fmt.Errorf("%w: %q", ErrParentNotFound, parentID)
	}

	return err
}
