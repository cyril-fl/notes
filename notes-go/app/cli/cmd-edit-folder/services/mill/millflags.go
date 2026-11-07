package millflagseditfolder

import (
	"context"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/spf13/pflag"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/process/editfolder"
	"main.go/app/shared/services/process/readitem"
)

type mill struct {
	plumprimitive.Mill
	cache

	ctx   context.Context
	db    bun.IDB
	flags *pflag.FlagSet
	id    string
}

type cache struct {
	input editfolder.Input
}

type Template struct {
	Ctx   context.Context
	Db    bun.IDB
	Flags *pflag.FlagSet
	ID    string
}

func New(params Template) *mill {
	return &mill{
		ctx:   params.Ctx,
		db:    params.Db,
		flags: params.Flags,
		id:    params.ID,
	}
}

// Resolve starts from the current folder, applies --title, or asks for it in a pre-filled form.
func (m *mill) Resolve() editfolder.Input {
	m.loadFolder()
	m.readFlags()
	m.askMissing()
	return m.cache.input
}

// Methods

func (m *mill) loadFolder() {
	process := readitem.Process(m.ctx, m.db, domain.FOLDER, m.id)
	process.Run()
	current := process.GetResult()
	m.cache.input = editfolder.Input{ID: current.Item.ID, Title: string(current.Item.Title)}
	m.SetErr(process.GetErr())
}

func (m *mill) readFlags() {
	if m.HasErr() || !m.flags.Changed("title") {
		return
	}

	title, err := m.flags.GetString("title")
	m.cache.input.Title = title
	m.SetErr(err)
}
