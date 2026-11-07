package encryptnotes

import (
	"context"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/process/linknotes"
	"main.go/app/shared/services/secret"
)

// Result is how many notes and accounts had text in clear.
type Result struct {
	Notes    int
	Accounts int
}

type run struct {
	plumprimitive.Process
	cache

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
}

type cache struct {
	result Result
}

// Process seals what every account wrote before encryption: the text of its notes and its snippets. It links those
// notes to their images first, which purgefiles finds by notes_on_files once the text is sealed. For the admin, once:
// a second run finds nothing left.
func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

func (p *run) Run() {
	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if err := p.sealNotes(ctx, tx); err != nil {
			return err
		}
		return p.sealSnippets(ctx, tx)
	}))
}

// Getters & Setters

func (p *run) GetResult() Result {
	return p.cache.result
}

// Methods

func (p *run) sealNotes(ctx context.Context, tx bun.Tx) error {
	var notes []domain.MetadataNote
	err := tx.NewSelect().Model(&notes).Column("item_id", "content").Where("content NOT LIKE ?", secret.Prefix+"%").Scan(ctx)
	if err != nil {
		return err
	}
	for _, note := range notes {
		if err := linknotes.Files(ctx, tx, note.ItemID, string(note.Content)); err != nil {
			return err
		}
		if _, err := tx.NewUpdate().Model(&note).Column("content").WherePK().Exec(ctx); err != nil {
			return err
		}
	}
	p.cache.result.Notes = len(notes)
	return nil
}

// sealSnippets rewrites the snippets kept as a JSON array: sealed, they are a JSON string.
func (p *run) sealSnippets(ctx context.Context, tx bun.Tx) error {
	var users []domain.User
	err := tx.NewSelect().Model(&users).Column("id", "snippets").Where("jsonb_typeof(snippets) = 'array'").Scan(ctx)
	if err != nil {
		return err
	}
	for _, user := range users {
		if _, err := tx.NewUpdate().Model(&user).Column("snippets").WherePK().Exec(ctx); err != nil {
			return err
		}
	}
	p.cache.result.Accounts = len(users)
	return nil
}
