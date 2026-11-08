package resetpassword

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/password"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/adduser"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

var ErrNotFound = errors.New("account not found")

type run struct {
	plumprimitive.Process

	ctx      context.Context
	db       bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	name     string
	password string
}

// Process gives an account a new password and logs all its devices out: for the admin, on the server.
func Process(ctx context.Context, db bun.IDB, name, newPassword string) *run {
	return &run{
		ctx:      ctx,
		db:       db,
		name:     strings.TrimSpace(name),
		password: newPassword,
	}
}

func (p *run) Run() {
	if len([]rune(p.password)) < adduser.MinPassword {
		p.SetErr(adduser.ErrShortPassword)
		return
	}

	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var id string
		err := tx.NewUpdate().Model((*domain.User)(nil)).
			Set("password_hash = ?", password.Hash(p.password)).
			Where("lower(name) = lower(?)", p.name).
			Returning("id").
			Scan(ctx, &id)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %q", ErrNotFound, p.name)
		}
		if err != nil {
			return err
		}
		_, err = tx.NewDelete().Model((*domain.Session)(nil)).Where("user_id = ?", id).Exec(ctx)
		return err
	}))
}
