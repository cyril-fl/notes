package millflagscreatefolder

import (
	"context"
	"errors"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/spf13/pflag"
	"github.com/uptrace/bun"
	"main.go/app/shared/services/process/createfolder"
)

var ErrNotImplemented = errors.New("--parent-from-list is not implemented yet: use --parent <uuid>")

type mill struct {
	plumprimitive.Mill
	cache

	ctx   context.Context
	db    bun.IDB
	flags *pflag.FlagSet
}

type cache struct {
	input          createfolder.Input
	parentFromList bool
}

type Template struct {
	Ctx   context.Context
	Db    bun.IDB
	Flags *pflag.FlagSet
}

func New(params Template) *mill {
	return &mill{
		ctx:   params.Ctx,
		db:    params.Db,
		flags: params.Flags,
	}
}

// Resolve turns the flags into a complete folder, asking in a form for what is missing.
func (m *mill) Resolve() createfolder.Input {
	m.readFlags()
	m.checkParentFromList()
	m.askMissing()

	return m.cache.input
}

// Methods

func (m *mill) readFlags() {
	var errs [3]error

	m.cache.input.Title, errs[0] = m.flags.GetString("title")
	m.cache.input.ParentID, errs[1] = m.flags.GetString("parent")
	m.cache.parentFromList, errs[2] = m.flags.GetBool("parent-from-list")

	m.SetErr(errors.Join(errs[:]...))
}

func (m *mill) checkParentFromList() {
	if m.HasErr() {
		return
	}

	if m.cache.parentFromList {
		m.SetErr(ErrNotImplemented)
	}
}
