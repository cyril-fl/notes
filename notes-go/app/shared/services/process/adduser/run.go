package adduser

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/password"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"
)

// MinPassword is the shortest password accepted, in runes.
const MinPassword = 8

var (
	ErrNoName        = errors.New("no account name given")
	ErrShortPassword = fmt.Errorf("the password needs %d characters or more", MinPassword)
	ErrNameTaken     = errors.New("an account with this name already exists")
	ErrBadInvite     = errors.New("this invitation is used, expired or unknown")
)

// Input is an account as the admin or the invited person gives it.
type Input struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	// The token of an invitation to consume, none when empty: the CLI, an admin, or the first account of a base.
	Invite string `json:"invite" binding:"optional"`
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
	// In a transaction, a savepoint: a name already taken must not abort the caller's. The invitation is consumed
	// in it too, so a name refused gives the link back instead of burning it.
	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if err := p.useInvite(ctx, tx); err != nil {
			return err
		}
		// The first account of a base runs the server: it creates the next ones.
		taken, err := tx.NewSelect().Model((*domain.User)(nil)).Exists(ctx)
		if err != nil {
			return err
		}
		p.cache.user.Admin = !taken

		var pgErr pgdriver.Error
		err = tx.NewInsert().Model(&p.cache.user).Returning("id, created_at").Scan(ctx)
		if errors.As(err, &pgErr) && pgErr.Field('C') == "23505" {
			return fmt.Errorf("%w: %q", ErrNameTaken, p.cache.user.Name)
		}
		return err
	}))
}

// useInvite marks the invitation used, and fails when another account took it already, or when it expired:
// the one UPDATE is the check, so two people cannot use the same link at once.
func (p *run) useInvite(ctx context.Context, tx bun.Tx) error {
	if p.input.Invite == "" {
		return nil
	}

	result, err := tx.NewUpdate().Model((*domain.Invite)(nil)).Set("used_at = current_timestamp").
		Where("token_hash = ? AND used_at IS NULL AND expires_at > current_timestamp", password.TokenHash(p.input.Invite)).
		Exec(ctx)
	if err != nil {
		return err
	}
	used, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if used != 1 {
		return ErrBadInvite
	}
	return nil
}
