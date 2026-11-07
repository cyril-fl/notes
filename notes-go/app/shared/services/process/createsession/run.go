package createsession

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/password"
)

// Lifetime is how long a session lasts without use: each use starts it again (readsession).
const Lifetime = 30 * 24 * time.Hour

var ErrBadCredentials = errors.New("wrong account name or password")

// dummyHash is checked when the name is unknown: the answer takes as long, and does not tell names apart.
var dummyHash = password.Hash("not a password")

type Input struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

// Result is the new session: the token goes to the client, and only there.
type Result struct {
	Token     string      `json:"-"`
	User      domain.User `json:"user"`
	ExpiresAt time.Time   `json:"expires_at"`
}

type run struct {
	plumprimitive.Process
	cache

	ctx   context.Context
	db    bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	input Input
}

type cache struct {
	result Result
}

// Process logs in: the account's password checked, a session made for the device.
func Process(ctx context.Context, db bun.IDB, input Input) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		input: input,
	}
}

func (p *run) Run() {
	p.checkPassword()
	p.saveSession()
}

// Getters & Setters

func (p *run) GetResult() Result {
	return p.cache.result
}

// Methods

func (p *run) checkPassword() {
	user := &p.cache.result.User
	err := p.db.NewSelect().Model(user).Where("lower(name) = lower(?)", strings.TrimSpace(p.input.Name)).Scan(p.ctx)
	if errors.Is(err, sql.ErrNoRows) {
		password.Verify(p.input.Password, dummyHash)
		p.SetErr(ErrBadCredentials)
		return
	}
	if err != nil {
		p.SetErr(err)
		return
	}
	if !password.Verify(p.input.Password, user.PasswordHash) {
		p.SetErr(ErrBadCredentials)
	}
}

func (p *run) saveSession() {
	if p.HasErr() {
		return
	}

	p.cache.result.Token = password.Token()
	p.cache.result.ExpiresAt = time.Now().Add(Lifetime)
	session := domain.Session{UserID: p.cache.result.User.ID, TokenHash: password.TokenHash(p.cache.result.Token), ExpiresAt: p.cache.result.ExpiresAt}
	_, err := p.db.NewInsert().Model(&session).Exec(p.ctx)
	p.SetErr(err)
}
