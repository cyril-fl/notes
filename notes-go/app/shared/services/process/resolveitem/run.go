package resolveitem

import (
	"context"
	"errors"
	"fmt"
	"strings"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
)

var (
	ErrNotFound  = errors.New("no item at this path")
	ErrAmbiguous = errors.New("several items at this path, use a uuid")
)

type run struct {
	plumprimitive.Process
	cache

	ctx  context.Context
	db   bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	kind domain.Kind
	refs []string
}

type cache struct {
	result []string
}

// Process turns references into ids: a path (/Projects/Go/Ideas) is looked up among the live items
// of the kind (any kind when empty), anything else is left for the caller to treat as a uuid.
func Process(ctx context.Context, db bun.IDB, kind domain.Kind, refs []string) *run {
	return &run{
		ctx:  ctx,
		db:   db,
		kind: kind,
		refs: refs,
	}
}

func (p *run) Run() {
	p.resolveRefs()
}

// Getters & Setters

// GetResult returns one id per reference, in the same order.
func (p *run) GetResult() []string {
	return p.cache.result
}

// Methods

func (p *run) resolveRefs() {
	for _, ref := range p.refs {
		if p.HasErr() {
			return
		}

		if !strings.HasPrefix(ref, "/") {
			p.cache.result = append(p.cache.result, ref)
			continue
		}
		p.resolvePath(ref)
	}
}

// resolvePath fails on several matches: titles are not unique, and acting on the wrong note is worse than an error.
func (p *run) resolvePath(path string) {
	paths, err := domain.Paths(p.ctx, p.db)
	if err != nil {
		p.SetErr(err)
		return
	}
	var ids []string
	for id, at := range paths {
		if at == path {
			ids = append(ids, id)
		}
	}
	if len(ids) > 0 {
		err = p.db.NewRaw("SELECT id FROM items WHERE id IN (?) AND (? = '' OR kind = ?) ORDER BY id", bun.In(ids), p.kind, p.kind).Scan(p.ctx, &ids)
	}
	if err != nil {
		p.SetErr(err)
		return
	}

	switch len(ids) {
	case 0:
		p.SetErr(fmt.Errorf("%w: %q", ErrNotFound, path))
	case 1:
		p.cache.result = append(p.cache.result, ids[0])
	default:
		p.SetErr(fmt.Errorf("%w: %q is %s", ErrAmbiguous, path, strings.Join(ids, ", ")))
	}
}
