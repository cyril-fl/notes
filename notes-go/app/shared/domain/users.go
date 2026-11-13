package domain

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	"github.com/uptrace/bun"
)

// User is an account: every item and tag belongs to one.
type User struct {
	bun.BaseModel `bun:"table:users,alias:u" json:"-"`

	ID           string `bun:",pk,type:uuid,default:gen_random_uuid()" json:"id"`
	Name         string `bun:",notnull" json:"name"` // unique case-insensitive, see createschema
	PasswordHash string `bun:",notnull" json:"-"`
	DataKey      string `bun:",nullzero" json:"-"`                          // the key of its notes sealed by NOTES_SECRET, see services/secret
	Locale       string `bun:",notnull,default:'fr'" json:"locale"`         // the language of the web app, one of Locales
	Theme        string `bun:",notnull,default:'notes'" json:"theme"`       // the colors of the web app, one of Themes
	ColorMode    string `bun:",notnull,default:'system'" json:"color_mode"` // light or dark, one of ColorModes
	View         string `bun:",notnull,default:'list'" json:"view"`         // how the web app lists items, one of Views
	Plan         string `bun:",notnull,default:'base'" json:"plan"`         // the features of the web app, one of Plans
	Admin        bool   `bun:",notnull,default:false" json:"admin"`         // may create, invite and delete accounts; the first account of the base gets it
	// Days before the trash is purged, 0 to never purge. An insert of 0 gets the default: only an update sets it.
	TrashRetentionDays int       `bun:",notnull,default:30" json:"trash_retention_days"`
	Avatar             string    `bun:",type:uuid,nullzero" json:"avatar,omitempty"` // the id of a File of the account, purgefiles keeps it
	Snippets           Snippets  `bun:",type:jsonb,notnull,default:'[]'" json:"snippets"`
	CreatedAt          time.Time `bun:",notnull,default:current_timestamp" json:"created_at"`
}

// Snippet is a piece of markdown of the account, inserted while typing or starting a new note; its #tags become tags of that note.
type Snippet struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// Snippets are kept sealed, as one JSON string in their jsonb column: write Seal's.
type Snippets []Snippet

// Seal gives the snippets sealed by the key of the account, as the JSON string to write.
func (s Snippets) Seal(userID string) (string, error) {
	if s == nil {
		s = Snippets{}
	}
	data, err := json.Marshal([]Snippet(s))
	if err != nil {
		return "", err
	}
	sealed, err := secret.Seal(userID, string(data))
	if err != nil {
		return "", err
	}
	data, err = json.Marshal(sealed)
	return string(data), err
}

// Scan also reads a JSON array in clear: the column's default, [].
func (s *Snippets) Scan(src any) error {
	var data []byte
	switch src := src.(type) {
	case string:
		data = []byte(src)
	case []byte:
		data = src
	}
	var sealed string
	if json.Unmarshal(data, &sealed) == nil {
		text, err := secret.Open(sealed)
		if err != nil {
			return err
		}
		data = []byte(text)
	}
	return json.Unmarshal(data, (*[]Snippet)(s))
}

// Locales are the languages of the web app, the first one by default.
var Locales = []string{"fr", "en"}

// Themes, ColorModes, Views and Plans are the other settings of the web app, the first one by default.
// Themes repeats the ids of notes-app/app/utils/themes.ts and Plans the plans of notes-app/app/utils/features.ts, users_test.go checks they match.
var (
	Themes     = []string{"notes", "dracula", "catppuccin", "nord", "gruvbox", "everforest", "tokyo", "vue"}
	ColorModes = []string{"system", "light", "dark"}
	Views      = []string{"list", "grid"}
	Plans      = []string{"base", "pro"}
)

// Session is a login of a device: the client keeps the token, the base only its hash.
type Session struct {
	bun.BaseModel `bun:"table:sessions,alias:s"`

	ID        string    `bun:",pk,type:uuid,default:gen_random_uuid()"`
	UserID    string    `bun:",notnull,type:uuid"`
	TokenHash string    `bun:",notnull,unique"`
	CreatedAt time.Time `bun:",notnull,default:current_timestamp"`
	ExpiresAt time.Time `bun:",notnull"`
}

// Invite is a link an admin hands out so someone creates their own account: the link holds the token, the base only its hash.
type Invite struct {
	bun.BaseModel `bun:"table:invites,alias:iv" json:"-"`

	ID        string    `bun:",pk,type:uuid,default:gen_random_uuid()" json:"id"`
	TokenHash string    `bun:",notnull,unique" json:"-"`
	UserID    string    `bun:",notnull,type:uuid" json:"created_by"` // the admin who made it
	CreatedAt time.Time `bun:",notnull,default:current_timestamp" json:"created_at"`
	ExpiresAt time.Time `bun:",notnull" json:"expires_at"`
	UsedAt    time.Time `bun:",nullzero" json:"-"` // set when it created an account, so a link serves once
}

// AddUserModes are the ways an account is created on a server, the first one by default: by invitation of an admin
// only, or by anyone reaching the web app.
var AddUserModes = []string{"invitation", "public"}

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
