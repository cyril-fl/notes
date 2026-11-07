package domain

import (
	"regexp"
	"time"

	"github.com/uptrace/bun"
	"main.go/app/shared/services/secret"
)

// File is an image of a note: its bytes are in Silo under Key(), the notes cite it as FileURL(ID).
type File struct {
	bun.BaseModel `bun:"table:files,alias:f" json:"-"`

	ID          string      `bun:",pk,type:uuid,default:gen_random_uuid()" json:"id"`
	UserID      string      `bun:",type:uuid,notnull,nullzero" json:"-"`
	Name        secret.Text `bun:",type:text" json:"name" swaggertype:"string"` // "" (NULL) for the images uploaded before names were sealed
	ContentType string      `bun:",notnull" json:"content_type"`
	Size        int64       `bun:",notnull" json:"size"`
	CreatedAt   time.Time   `bun:",notnull,default:current_timestamp" json:"created_at"`
	OrphanedAt  time.Time   `bun:",nullzero" json:"-"` // since no note cites it; purgefiles sets it
}

// FilePattern finds the files a note cites: the uuid is its first group.
var FilePattern = regexp.MustCompile(`/api/files/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`)

// FileURL is how a note cites a file.
func FileURL(id string) string {
	return "/api/files/" + id
}

// Key is where Silo keeps the bytes: under the account, so a key never reaches another one.
func (f File) Key() string {
	return f.UserID + "/" + f.ID
}

// FileTypes are the images accepted, with the extension of their export. Not SVG: it can hold scripts.
var FileTypes = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}
