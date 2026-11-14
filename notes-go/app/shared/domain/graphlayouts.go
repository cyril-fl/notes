package domain

import "github.com/uptrace/bun"

// GraphLayout is the settled layout of the whole graph of an account, computed by the server and shared by its devices.
// Data is sealed (secret.SealBytes): the tags are nodes by name, and the signature is a hash of them.
type GraphLayout struct {
	bun.BaseModel `bun:"table:graph_layouts,alias:gl" json:"-"`

	UserID string `bun:",pk,type:uuid"`
	Data   []byte `bun:",notnull,type:bytea"`
}
