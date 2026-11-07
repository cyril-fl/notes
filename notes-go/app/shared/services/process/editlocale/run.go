package editlocale

import (
	"context"
	"errors"
	"fmt"
	"slices"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
)

var ErrInvalidLocale = errors.New("unknown locale")

type run struct {
	plumprimitive.Process
	cache

	ctx    context.Context
	db     bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	locale string
}

type cache struct {
	result domain.User
}

func Process(ctx context.Context, db bun.IDB, locale string) *run {
	return &run{
		ctx:    ctx,
		db:     db,
		locale: locale,
	}
}

func (p *run) Run() {
	p.checkLocale()
	p.updateUser()
}

// Getters & Setters

// GetResult returns the account with its new locale.
func (p *run) GetResult() domain.User {
	return p.cache.result
}

// Methods
func (p *run) checkLocale() {
	if !slices.Contains(domain.Locales, p.locale) {
		p.SetErr(fmt.Errorf("%w: %q", ErrInvalidLocale, p.locale))
	}
}

func (p *run) updateUser() {
	if p.HasErr() {
		return
	}

	err := p.db.NewUpdate().
		Model(&p.cache.result).
		Set("locale = ?", p.locale).
		Where("id = ?", domain.UserID(p.ctx)).
		Returning("*").
		Scan(p.ctx)
	p.SetErr(err)
}
