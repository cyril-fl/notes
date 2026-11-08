package domain

import (
	"context"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	"github.com/uptrace/bun"
)

type Tag struct {
	bun.BaseModel `bun:"table:tags,alias:t" json:"-" yaml:"-"`

	ID        string      `bun:",pk,type:uuid,default:gen_random_uuid()" json:"id" yaml:"id"`
	UserID    string      `bun:",type:uuid,notnull,nullzero" json:"-" yaml:"-"`
	Value     secret.Text `bun:",type:text,notnull" json:"value" yaml:"value" swaggertype:"string"`
	ValueHash string      `bun:",notnull" json:"-" yaml:"-"` // its Blind, unique per account, see createschema
}

// SealTag gives a tag of the account of ctx to insert, its value sealed.
func SealTag(ctx context.Context, value string) (Tag, error) {
	sealed, err := secret.Seal(UserID(ctx), value)
	if err != nil {
		return Tag{}, err
	}
	hash, err := secret.Blind(UserID(ctx), value)
	return Tag{UserID: UserID(ctx), Value: sealed, ValueHash: hash}, err
}

// OpenTags opens the tag values a query gave sealed (an array_agg), ordered whatever their case.
func OpenTags(sealed []string) ([]string, error) {
	values := make([]string, len(sealed))
	for i, value := range sealed {
		text, err := secret.Open(value)
		if err != nil {
			return nil, err
		}
		values[i] = text
	}
	SortFold(values, func(v string) string { return v }, nil)
	return values, nil
}

// BlindTags gives the value_hash of each value: tags are found by it, whatever their case.
func BlindTags(ctx context.Context, values []string) ([]string, error) {
	hashes := make([]string, len(values))
	for i, value := range values {
		hash, err := secret.Blind(UserID(ctx), value)
		if err != nil {
			return nil, err
		}
		hashes[i] = hash
	}
	return hashes, nil
}

type NoteTag struct {
	bun.BaseModel `bun:"table:notes_on_tags,alias:nt"`

	TagID  string        `bun:",pk,type:uuid"`
	Tag    *Tag          `bun:"rel:belongs-to,join:tag_id=id"`
	NoteID string        `bun:",pk,type:uuid"`
	Note   *MetadataNote `bun:"rel:belongs-to,join:note_id=item_id"`
}

// NoteLink is a [[link]] of a note's text that names another note, resolved to its id.
type NoteLink struct {
	bun.BaseModel `bun:"table:notes_on_links,alias:nl"`

	SourceID string `bun:",pk,type:uuid"`
	TargetID string `bun:",pk,type:uuid"`
}

// NoteFile is an image a note cites: purgefiles reads it, the text of the note being sealed.
type NoteFile struct {
	bun.BaseModel `bun:"table:notes_on_files,alias:nf"`

	NoteID string `bun:",pk,type:uuid"`
	FileID string `bun:",pk,type:uuid"`
}
