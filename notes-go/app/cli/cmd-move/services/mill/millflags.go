package millflagsmove

import (
	"errors"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/spf13/pflag"
	"main.go/app/shared/services/process/moveitems"
)

var ErrNotImplemented = errors.New("--parent-from-list is not implemented yet: use --parent <uuid>")

type mill struct {
	plumprimitive.Mill
	cache

	flags *pflag.FlagSet
	ids   []string
}

type cache struct {
	input          moveitems.Input
	parentFromList bool
}

type Template struct {
	Flags *pflag.FlagSet
	IDs   []string
}

func New(params Template) *mill {
	return &mill{
		flags: params.Flags,
		ids:   params.IDs,
	}
}

// Resolve turns the args and flags into the move; no --parent means the root.
func (m *mill) Resolve() moveitems.Input {
	m.readFlags()
	m.checkParentFromList()
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

func (m *mill) checkParentFromList() {
	if m.HasErr() {
		return
	}

	if m.cache.parentFromList {
		m.SetErr(ErrNotImplemented)
	}
}
