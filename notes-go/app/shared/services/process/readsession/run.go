package readsession

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/password"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createsession"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

var ErrNoSession = errors.New("not logged in, or the session expired")

type run struct {
	plumprimitive.Process
	cache

	ctx   context.Context
	db    bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	token string
}

type cache struct {
	user domain.User
}

// Process finds the account of a session token, and starts its lifetime again: a device in use stays logged in.
// It loads the key of the account: its notes open from then on.
func Process(ctx context.Context, db bun.IDB, token string) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		token: token,
	}
}

func (p *run) Run() {
	if p.token == "" {
		p.SetErr(ErrNoSession)
		return
	}

	err := p.db.NewRaw(`
		WITH session AS (
			UPDATE sessions SET expires_at = ? WHERE token_hash = ? AND expires_at > current_timestamp RETURNING user_id
		)
		SELECT u.id, u.name, u.created_at FROM users AS u JOIN session ON session.user_id = u.id`,
		time.Now().Add(createsession.Lifetime), password.TokenHash(p.token)).
		Scan(p.ctx, &p.cache.user)
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrNoSession
	}
	if err == nil {
		err = secret.Load(p.ctx, p.db, p.cache.user.ID)
	}
	p.SetErr(err)
}

// Getters & Setters

func (p *run) GetResult() domain.User {
	return p.cache.user
}
