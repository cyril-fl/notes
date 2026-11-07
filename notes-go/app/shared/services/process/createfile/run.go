package createfile

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/secret"
	"main.go/app/shared/services/silo"
)

var (
	ErrNoData   = errors.New("no file given")
	ErrNotImage = errors.New("not a PNG, JPEG, GIF or WebP image")
)

// Input is a file as sent: its type is read from its bytes, not from what the sender claims.
type Input struct {
	Name string
	Data []byte
}

type run struct {
	plumprimitive.Process
	cache

	ctx   context.Context
	db    bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	store silo.Store
	input Input
}

type cache struct {
	file domain.File
}

// Process stores an image of the account: its row in Postgres, its bytes in Silo.
func Process(ctx context.Context, db bun.IDB, store silo.Store, input Input) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		store: store,
		input: input,
	}
}

func (p *run) Run() {
	p.makeFile()
	p.saveFile()
}

// Getters & Setters

// GetResult returns the created file, with its generated ID.
func (p *run) GetResult() domain.File {
	return p.cache.file
}

// Methods

func (p *run) makeFile() {
	if len(p.input.Data) == 0 {
		p.SetErr(ErrNoData)
		return
	}

	contentType := http.DetectContentType(p.input.Data)
	if _, ok := domain.FileTypes[contentType]; !ok {
		p.SetErr(fmt.Errorf("%w: %s", ErrNotImage, contentType))
		return
	}

	name := strings.TrimSpace(path.Base(strings.ReplaceAll(p.input.Name, "\\", "/")))
	if name == "" || name == "." || name == "/" {
		name = "image" + domain.FileTypes[contentType]
	}
	sealed, err := secret.Seal(domain.UserID(p.ctx), name)
	p.cache.file = domain.File{UserID: domain.UserID(p.ctx), Name: sealed, ContentType: contentType, Size: int64(len(p.input.Data))}
	p.SetErr(err)
}

// saveFile puts the bytes inside the transaction of the row: a failed upload leaves no row. A failed commit
// would leave an object without row, which nothing cites: its bytes are removed.
func (p *run) saveFile() {
	if p.HasErr() {
		return
	}

	put := false
	err := p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if err := tx.NewInsert().Model(&p.cache.file).Returning("*").Scan(ctx); err != nil {
			return err
		}
		put = true
		return p.store.Put(ctx, p.cache.file.Key(), p.input.Data, p.cache.file.ContentType)
	})
	if err != nil && put {
		_ = p.store.Remove(p.ctx, p.cache.file.Key())
	}
	p.SetErr(err)
}
