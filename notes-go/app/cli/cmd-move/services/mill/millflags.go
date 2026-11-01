package millflagsmove

import (
	"context"
	"errors"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/spf13/pflag"
	"github.com/uptrace/bun"
	"main.go/app/shared/services/huhform"
	"main.go/app/shared/services/process/listfolders"
	"main.go/app/shared/services/process/moveitems"
)

type mill struct {
	plumprimitive.Mill
	cache

	ctx   context.Context
	db    bun.IDB
	flags *pflag.FlagSet
	ids   []string
}

type cache struct {
	input          moveitems.Input
	parentFromList bool
}

type Template struct {
	Ctx   context.Context
	Db    bun.IDB
	Flags *pflag.FlagSet
	IDs   []string
}

func New(params Template) *mill {
	return &mill{
		ctx:   params.Ctx,
		db:    params.Db,
		flags: params.Flags,
		ids:   params.IDs,
	}
}

// Resolve turns the args and flags into the move; no --parent means the root.
func (m *mill) Resolve() moveitems.Input {
	m.readFlags()
	m.pickParent()
	return m.cache.input
}

// Methods

func (m *mill) readFlags() {
	var errs [2]error
	m.cache.input.IDs = m.ids
	m.cache.input.ParentID, errs[0] = m.flags.GetString("parent")
	m.cache.parentFromList, errs[1] = m.flags.GetBool("parent-from-list")
	m.SetErr(errors.Join(errs[:]...))
}

// pickParent asks for the parent in the folder list with --parent-from-list.
func (m *mill) pickParent() {
	if m.HasErr() || !m.cache.parentFromList {
		return
	}

	process := listfolders.Process(m.ctx, m.db)
	process.Run()
	if process.HasErr() {
		m.SetErr(process.GetErr())
		return
	}

	id, err := huhform.PickFolder(process.GetResult())
	m.cache.input.ParentID = id
	m.SetErr(err)
}
