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
	var notes []struct {
		ID      string `bun:"id"`
		UserID  string `bun:"user_id"`
		Content string `bun:"content"`
	}
	err := tx.NewRaw(`
		SELECT mn.item_id AS id, i.user_id, mn.content FROM metadata_notes AS mn JOIN items AS i ON i.id = mn.item_id
		WHERE mn.content NOT LIKE ?`, secret.Prefix+"%").Scan(ctx, &notes)
	if err != nil {
		return err
	}
	for _, note := range notes {
		if err := linknotes.Files(ctx, tx, note.ID, note.Content); err != nil {
			return err
		}
		if err := secret.Load(ctx, tx, note.UserID); err != nil {
			return err
		}
		sealed, err := secret.Seal(note.UserID, note.Content)
		if err != nil {
			return err
		}
		if _, err := tx.NewUpdate().Model((*domain.MetadataNote)(nil)).Set("content = ?", sealed).Where("item_id = ?", note.ID).Exec(ctx); err != nil {
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
		if err := secret.Load(ctx, tx, user.ID); err != nil {
			return err
		}
		sealed, err := user.Snippets.Seal(user.ID)
		if err != nil {
			return err
		}
		if _, err := tx.NewUpdate().Model((*domain.User)(nil)).Set("snippets = ?", sealed).Where("id = ?", user.ID).Exec(ctx); err != nil {
			return err
		}
	}
	p.cache.result.Accounts = len(users)
	return nil
}
