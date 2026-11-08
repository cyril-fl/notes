package deletesession

import (
	"context"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/password"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

type run struct {
	plumprimitive.Process

	ctx   context.Context
	db    bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	token string
}

// Process logs a device out: its session is deleted, an unknown token is no error.
func Process(ctx context.Context, db bun.IDB, token string) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		token: token,
	}
}

func (p *run) Run() {
	_, err := p.db.NewDelete().Model((*domain.Session)(nil)).Where("token_hash = ?", password.TokenHash(p.token)).Exec(p.ctx)
	p.SetErr(err)
}
