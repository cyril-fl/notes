// Package testaccount gives tests an account to act for, inside their rolled-back transaction.
package testaccount

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	"github.com/uptrace/bun"
)

// New creates an account in tx and returns ctx acting for it, with a savepoint of tx begun with that ctx: a bun.Tx
// hands its RunInTx functions the ctx it began with, so processes nested in a transaction keep the account.
// Its name is unique per run; no real password: hashing one would only slow the tests down.
func New(t *testing.T, ctx context.Context, tx bun.Tx) (context.Context, bun.Tx) {
	t.Helper()
	user := domain.User{Name: fmt.Sprintf("test-%d", time.Now().UnixNano()), PasswordHash: "-"}
	if err := tx.NewInsert().Model(&user).Returning("id").Scan(ctx); err != nil {
		t.Fatalf("test account: %v", err)
	}
	if err := secret.Load(ctx, tx, user.ID); err != nil {
		t.Fatalf("test account: %v", err)
	}
	ctx = domain.WithUser(ctx, user.ID)
	savepoint, err := tx.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("test account: %v", err)
	}
	return ctx, savepoint
}
