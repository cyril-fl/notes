package edituser

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/process/adduser"
)

// MaxRetentionDays keeps the trash retention a sane number of days, ten years.
const MaxRetentionDays = 3650

// MaxSnippets and MaxSnippetSize keep the snippets of an account a small list.
const (
	MaxSnippets    = 100
	MaxSnippetSize = 64 << 10
)

var (
	ErrInvalidLocale    = errors.New("unknown locale")
	ErrInvalidTheme     = errors.New("unknown theme")
	ErrInvalidColorMode = errors.New("unknown color mode")
	ErrInvalidView      = errors.New("unknown view")
	ErrInvalidRetention = fmt.Errorf("the trash retention is between 0 and %d days", MaxRetentionDays)
	ErrInvalidAvatar    = errors.New("the avatar is not an image of the account")
	ErrInvalidSnippet   = fmt.Errorf("a snippet needs a name, at most %d of them of %d bytes each", MaxSnippets, MaxSnippetSize)
)

// Input holds the settings to change: a nil one is kept.
type Input struct {
	Locale             *string           `json:"locale,omitempty" binding:"optional"`
	Theme              *string           `json:"theme,omitempty" binding:"optional"`
	ColorMode          *string           `json:"color_mode,omitempty" binding:"optional"`
	View               *string           `json:"view,omitempty" binding:"optional"`
	TrashRetentionDays *int              `json:"trash_retention_days,omitempty" binding:"optional"`
	Name               *string           `json:"name,omitempty" binding:"optional"`
	Avatar             *string           `json:"avatar,omitempty" binding:"optional"`   // the id of an image of the account, "" for none
	Snippets           *[]domain.Snippet `json:"snippets,omitempty" binding:"optional"` // the whole list, it replaces the saved one
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
	p.checkAvatar()
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
	case p.input.Name != nil && strings.TrimSpace(*p.input.Name) == "":
		p.SetErr(adduser.ErrNoName)
	case !validSnippets(p.input.Snippets):
		p.SetErr(ErrInvalidSnippet)
	}
}

func (p *run) checkAvatar() {
	if p.HasErr() || p.input.Avatar == nil || *p.input.Avatar == "" {
		return
	}

	// id::text: a value that is not a uuid is not found, rather than a failed query.
	found, err := p.db.NewSelect().Model((*domain.File)(nil)).
		Where("id::text = ? AND user_id = ?", *p.input.Avatar, domain.UserID(p.ctx)).
		Exists(p.ctx)
	if err == nil && !found {
		err = fmt.Errorf("%w: %q", ErrInvalidAvatar, *p.input.Avatar)
	}
	p.SetErr(err)
}

func (p *run) updateUser() {
	if p.HasErr() {
		return
	}

	// In a transaction, a savepoint: a name already taken must not abort the caller's.
	p.SetErr(p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// "id = id" keeps the query valid when nothing is set: it then only reads the account back.
		query := tx.NewUpdate().
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
		if p.input.Name != nil {
			query.Set("name = ?", strings.TrimSpace(*p.input.Name))
		}
		if p.input.Avatar != nil {
			query.Set("avatar = NULLIF(?, '')::uuid", *p.input.Avatar)
		}

		// shortcut: the last of two devices editing the snippets wins, per-snippet endpoints if that bites.
		if p.input.Snippets != nil {
			snippets := *p.input.Snippets
			if snippets == nil {
				snippets = []domain.Snippet{} // null would break the not null column
			}
			query.Set("snippets = ?", domain.Snippets(snippets))
		}

		var pgErr pgdriver.Error
		err := query.Scan(ctx)
		if errors.As(err, &pgErr) && pgErr.Field('C') == "23505" {
			return fmt.Errorf("%w: %q", adduser.ErrNameTaken, strings.TrimSpace(*p.input.Name))
		}
		return err
	}))
}

// Helpers

// valid accepts a setting left out, or one of choices.
func valid(value *string, choices []string) bool {
	return value == nil || slices.Contains(choices, *value)
}

// validSnippets accepts snippets left out, or a short list of named ones.
func validSnippets(snippets *[]domain.Snippet) bool {
	if snippets == nil {
		return true
	}
	if len(*snippets) > MaxSnippets {
		return false
	}
	for _, snippet := range *snippets {
		if strings.TrimSpace(snippet.Name) == "" || len(snippet.Content) > MaxSnippetSize {
			return false
		}
	}
	return true
}
