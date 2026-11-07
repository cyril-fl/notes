package purgeaccount

import (
	"context"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
)

type run struct {
	plumprimitive.Process

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
}

// Process deletes for good every note, folder, tag and image of the account, the trash included. The account and
// its settings stay.
func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

func (p *run) Run() {
	p.purge()
}

// Methods

// purge hard deletes the items, the ON DELETE CASCADE FKs remove metadata, tag links and links. The images get
// orphaned_at = -infinity, cited or not: purgefiles removes their bytes from Silo at its next round.
func (p *run) purge() {
	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		user := domain.UserID(ctx)
		if _, err := tx.NewDelete().Model((*domain.Item)(nil)).ForceDelete().Where("user_id = ?", user).Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewDelete().Model((*domain.Tag)(nil)).Where("user_id = ?", user).Exec(ctx); err != nil {
			return err
		}
		_, err := tx.NewUpdate().Model((*domain.File)(nil)).Set("orphaned_at = '-infinity'").Where("user_id = ?", user).Exec(ctx)
		return err
	}))
}
