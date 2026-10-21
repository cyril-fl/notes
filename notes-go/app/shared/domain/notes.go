package domain

import (
	"time"

	"github.com/uptrace/bun"
)

type Kind string

const (
	FOLDER Kind = "folder"
	NOTE   Kind = "note"
)

type Item struct {
	bun.BaseModel `bun:"table:items,alias:i"`

	ID        string    `bun:",pk,type:uuid,default:gen_random_uuid()"`
	ParentID  *string   `bun:",type:uuid,nullzero"`
	Parent    *Item     `bun:"rel:belongs-to,join:parent_id=id"`
	CreatedAt time.Time `bun:",notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:",nullzero"`
	DeletedAt time.Time `bun:",soft_delete,nullzero"`
	Kind      Kind      `bun:",notnull"`
	Title     string    `bun:"-"`
}

type Metadata struct {
	ItemID string `bun:",pk,type:uuid"`
	Item   *Item  `bun:"rel:belongs-to,join:item_id=id"`
}

type MetadataNote struct {
	bun.BaseModel `bun:"table:metadata_notes,alias:mn"`
	Metadata      `bun:",inline"`

	Content  string `bun:",type:text"`
	Readonly bool   `bun:",notnull"`
	Pinned   bool   `bun:",notnull"`
	Tags     []Tag  `bun:"m2m:notes_on_tags,join:Note=Tag"`
}

type MetadataFolder struct {
	bun.BaseModel `bun:"table:metadata_folders,alias:mf"`
	Metadata      `bun:",inline"`
}
