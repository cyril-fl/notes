package createfolder

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"
)

var (
	ErrNoTitle        = errors.New("no folder title given")
	ErrParentNotFound = errors.New("parent folder not found")
)

// foreignKeyViolation is the Postgres SQLSTATE raised here by items_parent_user_fkey (a missing parent, or another account's).
const foreignKeyViolation = "23503"

// Input is a folder as the user gave it: an empty ParentID puts it at the root.
type Input struct {
	Title    string `json:"title"`
	ParentID string `json:"parent_id" binding:"optional"`
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
}

func Process(ctx context.Context, db bun.IDB, input Input) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		input: input,
	}
}

func (p *run) Run() {
	p.makeFolder()
	p.saveFolder()
}

// Getters & Setters

// GetResult returns the created item, with its generated ID.
func (p *run) GetResult() domain.Item {
	return p.cache.item
}

// Methods

func (p *run) makeFolder() {
	p.cache.input = Input{
		Title:    strings.TrimSpace(p.input.Title),
		ParentID: strings.TrimSpace(p.input.ParentID),
	}

	if p.cache.input.Title == "" {
		p.SetErr(ErrNoTitle)
	}
}

// saveFolder writes item and metadata in one transaction: a failure leaves no half folder.
func (p *run) saveFolder() {
	if p.HasErr() {
		return
	}

	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, step := range []func(context.Context, bun.Tx) error{p.checkParent, p.insertItem, p.insertMetadata} {
			if err := step(ctx, tx); err != nil {
				return err
			}
		}
		return nil
	}))
}

// checkParent wants a folder not in the trash: the FK alone accepts a trashed one.
func (p *run) checkParent(ctx context.Context, tx bun.Tx) error {
	if p.cache.input.ParentID == "" {
		return nil
	}

	exists, err := tx.NewSelect().
		Model((*domain.Item)(nil)).
		Where("id = ? AND kind = ? AND user_id = ?", p.cache.input.ParentID, domain.FOLDER, domain.UserID(p.ctx)).
		Exists(ctx)
	if err == nil && !exists {
		err = fmt.Errorf("%w: %q", ErrParentNotFound, p.cache.input.ParentID)
	}
	return err
}

func (p *run) insertItem(ctx context.Context, tx bun.Tx) error {
	title, hash, err := domain.SealTitle(p.ctx, p.cache.input.Title)
	if err != nil {
		return err
	}
	p.cache.item = domain.Item{UserID: domain.UserID(p.ctx), Kind: domain.FOLDER, Title: title, TitleHash: hash}
	if p.cache.input.ParentID != "" {
		p.cache.item.ParentID = &p.cache.input.ParentID
	}

	// Not "*": parent_kind is a DB-only column, Item cannot scan it.
	err = tx.NewInsert().Model(&p.cache.item).Returning("id, created_at").Scan(ctx)
	p.cache.item.Title = secret.Text(p.cache.input.Title)
	return toDomainErr(err, p.cache.input.ParentID)
}

func (p *run) insertMetadata(ctx context.Context, tx bun.Tx) error {
	metadata := domain.MetadataFolder{Metadata: domain.Metadata{ItemID: p.cache.item.ID}}

	_, err := tx.NewInsert().Model(&metadata).Exec(ctx)
	return err
}

// Helpers

// toDomainErr turns the parent FK violation into ErrParentNotFound: a missing id, or one that is not a folder.
func toDomainErr(err error, parentID string) error {
	if domain.FolderNameTaken(err) {
		return domain.ErrFolderExists
	}
	var pgErr pgdriver.Error

	if errors.As(err, &pgErr) && pgErr.Field('C') == foreignKeyViolation {
		return fmt.Errorf("%w: %q", ErrParentNotFound, parentID)
	}

	return err
}
