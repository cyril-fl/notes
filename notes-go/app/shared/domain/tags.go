package domain

import "github.com/uptrace/bun"

type Tag struct {
	bun.BaseModel `bun:"table:tags,alias:t" json:"-" yaml:"-"`

	ID     string `bun:",pk,type:uuid,default:gen_random_uuid()" json:"id" yaml:"id"`
	UserID string `bun:",type:uuid,notnull,nullzero" json:"-" yaml:"-"`
	Value  string `bun:",notnull" json:"value" yaml:"value"` // unique case-insensitive, see CreateTables
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
