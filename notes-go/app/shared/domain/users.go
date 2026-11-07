package domain

import (
	"context"
	"errors"
	"time"

	"github.com/uptrace/bun"
)

// User is an account: every item and tag belongs to one.
type User struct {
	bun.BaseModel `bun:"table:users,alias:u" json:"-"`

	ID           string    `bun:",pk,type:uuid,default:gen_random_uuid()" json:"id"`
	Name         string    `bun:",notnull" json:"name"` // unique case-insensitive, see createschema
	PasswordHash string    `bun:",notnull" json:"-"`
	Locale       string    `bun:",notnull,default:'fr'" json:"locale"` // the language of the web app, one of Locales
	CreatedAt    time.Time `bun:",notnull,default:current_timestamp" json:"created_at"`
}

// Locales are the languages of the web app, the first one by default.
var Locales = []string{"fr", "en"}

// Session is a login of a device: the client keeps the token, the base only its hash.
type Session struct {
	bun.BaseModel `bun:"table:sessions,alias:s"`

	ID        string    `bun:",pk,type:uuid,default:gen_random_uuid()"`
	UserID    string    `bun:",notnull,type:uuid"`
	TokenHash string    `bun:",notnull,unique"`
	CreatedAt time.Time `bun:",notnull,default:current_timestamp"`
	ExpiresAt time.Time `bun:",notnull"`
}

var ErrNoUser = errors.New("not logged in")

type userKey struct{}

// WithUser gives ctx the account the processes act for: the CLI from its session, the API from the request's.
func WithUser(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userKey{}, userID)
}

// UserID is the account of ctx, "" without one. A bun.Tx hands its RunInTx functions the ctx it began with, not the
// one given: begin transactions with a ctx that has the account (the API and the CLI do, tests use testaccount).
func UserID(ctx context.Context) string {
	id, _ := ctx.Value(userKey{}).(string)
	return id
}
