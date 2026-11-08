package millflagscreatenote

import (
	"context"
	"errors"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/huhform"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listfolders"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listtag"
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
	input          createnote.Input
	tagFromList    bool
	parentFromList bool
	tags           []domain.Tag
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

// Resolve turns the flags into a complete note, asking in a form for what is missing.
func (m *mill) Resolve() createnote.Input {
	m.readFlags()
	m.pickParent()
	m.loadTags()
	m.askMissing()

	return m.cache.input
}

// Methods

func (m *mill) readFlags() {
	var errs [4]error

	m.cache.input.Content, errs[0] = m.flags.GetString("content")
	m.cache.input.ParentID, errs[1] = m.flags.GetString("parent")
	m.cache.input.Tags, errs[2] = m.flags.GetStringSlice("tag")
	errs[3] = m.readBoolFlags(map[string]*bool{
		"tag-from-list":    &m.cache.tagFromList,
		"parent-from-list": &m.cache.parentFromList,
	})

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

func (m *mill) loadTags() {
	if m.HasErr() || !m.cache.tagFromList {
		return
	}

	process := listtag.Process(m.ctx, m.db)
	process.Run()

	m.cache.tags = process.GetResult()
	m.SetErr(process.GetErr())
}

// Helpers

// readBoolFlags mirrors plumcliutilsargs.GetBoolFlags, which needs a *cobra.Command and not a FlagSet.
func (m *mill) readBoolFlags(targets map[string]*bool) error {
	for name, target := range targets {
		value, err := m.flags.GetBool(name)
		if err != nil {
			return err
		}
		*target = value
	}

	return nil
}
