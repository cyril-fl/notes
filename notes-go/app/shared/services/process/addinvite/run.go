package addinvite

import (
	"context"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/password"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

// Lifetime is how long an invitation stays usable: long enough to be sent and read, short enough that a forgotten
// link closes by itself.
const Lifetime = 7 * 24 * time.Hour

// Result is the new invitation: the token goes to the admin once, to put in the link, and is never readable again.
type Result struct {
	Token  string        `json:"token"`
	Invite domain.Invite `json:"invite"`
}

type run struct {
	plumprimitive.Process
	cache

	ctx context.Context
	db  bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
}

type cache struct {
	result Result
}

// Process creates an invitation for the admin of ctx.
func Process(ctx context.Context, db bun.IDB) *run {
	return &run{
		ctx: ctx,
		db:  db,
	}
}

func (p *run) Run() {
	p.saveInvite()
}

// Getters & Setters

func (p *run) GetResult() Result {
	return p.cache.result
}

// Methods

func (p *run) saveInvite() {
	p.cache.result.Token = password.Token()
	p.cache.result.Invite = domain.Invite{
		UserID:    domain.UserID(p.ctx),
		TokenHash: password.TokenHash(p.cache.result.Token),
		ExpiresAt: time.Now().Add(Lifetime),
	}
	_, err := p.db.NewInsert().Model(&p.cache.result.Invite).Returning("id, created_at").Exec(p.ctx)
	p.SetErr(err)
}
