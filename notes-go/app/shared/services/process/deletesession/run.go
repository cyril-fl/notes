package deletesession

import (
	"context"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/password"
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
