package millflagscreatefolder

import (
	"context"
	"errors"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/huhform"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createfolder"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listfolders"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/spf13/pflag"
	"github.com/uptrace/bun"
)

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
	m.pickParent()
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
