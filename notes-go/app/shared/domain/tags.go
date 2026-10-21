package domain

import "github.com/uptrace/bun"

type Tag struct {
	bun.BaseModel `bun:"table:tags,alias:t"`

	ID    string `bun:",pk,type:uuid,default:gen_random_uuid()"`
	Value string `bun:",notnull"` // unique case-insensitive, see CreateTables
}

type NoteTag struct {
	bun.BaseModel `bun:"table:notes_on_tags,alias:nt"`

	TagID  string        `bun:",pk,type:uuid"`
	Tag    *Tag          `bun:"rel:belongs-to,join:tag_id=id"`
	NoteID string        `bun:",pk,type:uuid"`
	Note   *MetadataNote `bun:"rel:belongs-to,join:note_id=item_id"`
}
