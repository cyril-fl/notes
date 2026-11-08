package editpassword

import (
	"context"
	"errors"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/password"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/adduser"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

var ErrWrongPassword = errors.New("the current password is wrong")

// Input is the password the account has, and the one it wants.
type Input struct {
	Current string `json:"current"`
	New     string `json:"new"`
}

type run struct {
	plumprimitive.Process

	ctx   context.Context
	db    bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	input Input
	token string
}

// Process gives the account of ctx a new password and logs its other devices out; the session of token stays.
func Process(ctx context.Context, db bun.IDB, input Input, token string) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		input: input,
		token: token,
	}
}

func (p *run) Run() {
	p.checkInput()
	p.checkPassword()
	p.savePassword()
}

// Methods

func (p *run) checkInput() {
	if len([]rune(p.input.New)) < adduser.MinPassword {
		p.SetErr(adduser.ErrShortPassword)
	}
}

func (p *run) checkPassword() {
	if p.HasErr() {
		return
	}

	var user domain.User
	err := p.db.NewSelect().Model(&user).Column("password_hash").Where("id = ?", domain.UserID(p.ctx)).Scan(p.ctx)
	if err == nil && !password.Verify(p.input.Current, user.PasswordHash) {
		err = ErrWrongPassword
	}
	p.SetErr(err)
}

func (p *run) savePassword() {
	if p.HasErr() {
		return
	}

	// The account of p.ctx: RunInTx hands a bun.Tx's own ctx, which may act for another.
	id := domain.UserID(p.ctx)
	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewUpdate().Model((*domain.User)(nil)).
			Set("password_hash = ?", password.Hash(p.input.New)).
			Where("id = ?", id).
			Exec(ctx)
		if err != nil {
			return err
		}
		_, err = tx.NewDelete().Model((*domain.Session)(nil)).
			Where("user_id = ? AND token_hash <> ?", id, password.TokenHash(p.token)).
			Exec(ctx)
		return err
	}))
}
