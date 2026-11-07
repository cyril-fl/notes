package adduser

import (
	"context"
	"errors"
	"fmt"
	"strings"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/password"
)

// MinPassword is the shortest password accepted, in runes.
const MinPassword = 8

var (
	ErrNoName        = errors.New("no account name given")
	ErrShortPassword = fmt.Errorf("the password needs %d characters or more", MinPassword)
	ErrNameTaken     = errors.New("an account with this name already exists")
)

// Input is an account as the admin gives it.
type Input struct {
	Name     string
	Password string
}

type run struct {
	plumprimitive.Process
	cache

	ctx   context.Context
	db    bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	input Input
}

type cache struct {
	user    domain.User
	adopted int64
}

// Process creates an account. The first one adopts what the base held before accounts: nothing is left without one.
func Process(ctx context.Context, db bun.IDB, input Input) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		input: input,
	}
}

func (p *run) Run() {
	p.checkInput()
	p.saveUser()
}

// Getters & Setters

func (p *run) GetResult() domain.User {
	return p.cache.user
}

// GetAdopted returns how many items and tags made before accounts the new account took.
func (p *run) GetAdopted() int64 {
	return p.cache.adopted
}

// Methods

func (p *run) checkInput() {
	switch {
	case strings.TrimSpace(p.input.Name) == "":
		p.SetErr(ErrNoName)
	case len([]rune(p.input.Password)) < MinPassword:
		p.SetErr(ErrShortPassword)
	}
}

func (p *run) saveUser() {
	if p.HasErr() {
		return
	}

	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		p.cache.user = domain.User{Name: strings.TrimSpace(p.input.Name), PasswordHash: password.Hash(p.input.Password)}
		var pgErr pgdriver.Error
		err := tx.NewInsert().Model(&p.cache.user).Returning("id, created_at").Scan(ctx)
		if errors.As(err, &pgErr) && pgErr.Field('C') == "23505" {
			return fmt.Errorf("%w: %q", ErrNameTaken, p.cache.user.Name)
		}
		if err != nil {
			return err
		}

		for _, table := range []string{"items", "tags"} {
			result, err := tx.NewRaw("UPDATE ? SET user_id = ? WHERE user_id IS NULL", bun.Ident(table), p.cache.user.ID).Exec(ctx)
			if err != nil {
				return err
			}
			rows, _ := result.RowsAffected()
			p.cache.adopted += rows
		}
		return nil
	}))
}
