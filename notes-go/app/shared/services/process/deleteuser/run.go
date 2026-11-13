package deleteuser

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/silo"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

var (
	ErrNotFound  = errors.New("no account with this id")
	ErrSelf      = errors.New("an account cannot delete itself")
	ErrLastAdmin = errors.New("the last admin account cannot be deleted")
)

type run struct {
	plumprimitive.Process
	cache

	ctx    context.Context
	db     bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	store  silo.Store
	userID string
}

type cache struct {
	user  domain.User
	files []domain.File
}

// Process deletes an account for good, with its notes, folders, tags, images and sessions: the admin of ctx asks it,
// for another account than its own.
func Process(ctx context.Context, db bun.IDB, store silo.Store, userID string) *run {
	return &run{
		ctx:    ctx,
		db:     db,
		store:  store,
		userID: userID,
	}
}

func (p *run) Run() {
	p.loadUser()
	p.loadFiles()
	p.removeBytes()
	p.deleteUser()
}

// Getters & Setters

// GetResult returns the account deleted.
func (p *run) GetResult() domain.User {
	return p.cache.user
}

// Methods

func (p *run) loadUser() {
	if p.userID == domain.UserID(p.ctx) {
		p.SetErr(ErrSelf)
		return
	}

	err := p.db.NewSelect().Model(&p.cache.user).ExcludeColumn("snippets").Where("id = ?", p.userID).Scan(p.ctx)
	if errors.Is(err, sql.ErrNoRows) {
		p.SetErr(ErrNotFound)
		return
	}
	if err != nil {
		p.SetErr(err)
		return
	}
	if p.cache.user.Admin {
		admins, err := p.db.NewSelect().Model((*domain.User)(nil)).Where("admin").Count(p.ctx)
		if err != nil {
			p.SetErr(err)
			return
		}
		if admins <= 1 {
			p.SetErr(ErrLastAdmin)
		}
	}
}

func (p *run) loadFiles() {
	if p.HasErr() {
		return
	}

	err := p.db.NewSelect().Model(&p.cache.files).Column("f.id", "f.user_id").Where("user_id = ?", p.userID).Scan(p.ctx)
	p.SetErr(err)
}

// removeBytes empties the account's images from Silo before its rows go: the cascade would take the files rows with
// them, and no purge starts from the bucket's side (purgefiles.purgeStrays stops when the table is empty).
func (p *run) removeBytes() {
	if p.HasErr() {
		return
	}

	var errs []error
	for _, file := range p.cache.files {
		if err := p.store.Remove(p.ctx, file.Key()); err != nil {
			errs = append(errs, err)
		}
	}
	p.SetErr(errors.Join(errs...))
}

// deleteUser removes the row; the ON DELETE CASCADE FKs take its items, tags, files, sessions and invitations.
func (p *run) deleteUser() {
	if p.HasErr() {
		return
	}

	_, err := p.db.NewDelete().Model((*domain.User)(nil)).Where("id = ?", p.userID).Exec(p.ctx)
	p.SetErr(err)
}
