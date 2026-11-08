package rotatesecret

import (
	"context"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

type run struct {
	plumprimitive.Process
	cache

	ctx       context.Context
	db        bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	oldSecret string
}

type cache struct {
	rewrapped int
}

// Process seals the key of every account, sealed by oldSecret, by NOTES_SECRET instead: the notes stay as they are.
// For the admin, after a new NOTES_SECRET; a second run finds nothing left. All or nothing: a key that neither
// secret opens stops it.
func Process(ctx context.Context, db bun.IDB, oldSecret string) *run {
	return &run{
		ctx:       ctx,
		db:        db,
		oldSecret: oldSecret,
	}
}

func (p *run) Run() {
	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var users []domain.User
		if err := tx.NewSelect().Model(&users).Column("id", "data_key").Where("data_key IS NOT NULL").Scan(ctx); err != nil {
			return err
		}
		for _, user := range users {
			wrapped, err := secret.Rewrap(user.DataKey, p.oldSecret)
			if err != nil {
				return err
			}
			if wrapped == user.DataKey {
				continue
			}
			if _, err := tx.NewUpdate().Model((*domain.User)(nil)).Set("data_key = ?", wrapped).Where("id = ?", user.ID).Exec(ctx); err != nil {
				return err
			}
			p.cache.rewrapped++
		}
		return nil
	}))
}

// Getters & Setters

// GetResult returns how many keys were sealed again.
func (p *run) GetResult() int {
	return p.cache.rewrapped
}
