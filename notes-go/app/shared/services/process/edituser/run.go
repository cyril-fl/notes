package edituser

import (
	"context"
	"errors"
	"fmt"
	"slices"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
)

// MaxRetentionDays keeps the trash retention a sane number of days, ten years.
const MaxRetentionDays = 3650

var (
	ErrInvalidLocale    = errors.New("unknown locale")
	ErrInvalidTheme     = errors.New("unknown theme")
	ErrInvalidColorMode = errors.New("unknown color mode")
	ErrInvalidView      = errors.New("unknown view")
	ErrInvalidRetention = fmt.Errorf("the trash retention is between 0 and %d days", MaxRetentionDays)
)

// Input holds the settings to change: a nil one is kept.
type Input struct {
	Locale             *string `json:"locale,omitempty" binding:"optional"`
	Theme              *string `json:"theme,omitempty" binding:"optional"`
	ColorMode          *string `json:"color_mode,omitempty" binding:"optional"`
	View               *string `json:"view,omitempty" binding:"optional"`
	TrashRetentionDays *int    `json:"trash_retention_days,omitempty" binding:"optional"`
}

type run struct {
	plumprimitive.Process
	cache

	ctx   context.Context
	db    bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	input Input
}

type cache struct {
	result domain.User
}

// Process changes the settings of the account of ctx.
func Process(ctx context.Context, db bun.IDB, input Input) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		input: input,
	}
}

func (p *run) Run() {
	p.checkInput()
	p.updateUser()
}

// Getters & Setters

// GetResult returns the account with its new settings.
func (p *run) GetResult() domain.User {
	return p.cache.result
}

// Methods
func (p *run) checkInput() {
	switch {
	case !valid(p.input.Locale, domain.Locales):
		p.SetErr(fmt.Errorf("%w: %q", ErrInvalidLocale, *p.input.Locale))
	case !valid(p.input.Theme, domain.Themes):
		p.SetErr(fmt.Errorf("%w: %q", ErrInvalidTheme, *p.input.Theme))
	case !valid(p.input.ColorMode, domain.ColorModes):
		p.SetErr(fmt.Errorf("%w: %q", ErrInvalidColorMode, *p.input.ColorMode))
	case !valid(p.input.View, domain.Views):
		p.SetErr(fmt.Errorf("%w: %q", ErrInvalidView, *p.input.View))
	case p.input.TrashRetentionDays != nil && (*p.input.TrashRetentionDays < 0 || *p.input.TrashRetentionDays > MaxRetentionDays):
		p.SetErr(fmt.Errorf("%w: %d", ErrInvalidRetention, *p.input.TrashRetentionDays))
	}
}

func (p *run) updateUser() {
	if p.HasErr() {
		return
	}

	// "id = id" keeps the query valid when nothing is set: it then only reads the account back.
	query := p.db.NewUpdate().
		Model(&p.cache.result).
		Set("id = id").
		Where("id = ?", domain.UserID(p.ctx)).
		Returning("*")
	if p.input.Locale != nil {
		query.Set("locale = ?", *p.input.Locale)
	}
	if p.input.Theme != nil {
		query.Set("theme = ?", *p.input.Theme)
	}
	if p.input.ColorMode != nil {
		query.Set("color_mode = ?", *p.input.ColorMode)
	}
	if p.input.View != nil {
		query.Set("view = ?", *p.input.View)
	}
	if p.input.TrashRetentionDays != nil {
		query.Set("trash_retention_days = ?", *p.input.TrashRetentionDays)
	}
	p.SetErr(query.Scan(p.ctx))
}

// Helpers

// valid accepts a setting left out, or one of choices.
func valid(value *string, choices []string) bool {
	return value == nil || slices.Contains(choices, *value)
}
