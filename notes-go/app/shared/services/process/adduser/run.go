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
	user domain.User
}

// Process creates an account.
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

	p.cache.user = domain.User{Name: strings.TrimSpace(p.input.Name), PasswordHash: password.Hash(p.input.Password)}
	// In a transaction, a savepoint: a name already taken must not abort the caller's.
	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var pgErr pgdriver.Error
		err := tx.NewInsert().Model(&p.cache.user).Returning("id, created_at").Scan(ctx)
		if errors.As(err, &pgErr) && pgErr.Field('C') == "23505" {
			return fmt.Errorf("%w: %q", ErrNameTaken, p.cache.user.Name)
		}
		return err
	}))
}
