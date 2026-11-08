package millflagseditnote

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/hashtag"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/editnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listtag"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
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
	id    string
}

type cache struct {
	input       editnote.Input
	content     string
	addTags     []string
	removeTags  []string
	tagFromList bool
	tags        []domain.Tag
	version     time.Time
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

// Resolve starts from the current note, applies the flags, and opens a form when no change was given.
func (m *mill) Resolve() editnote.Input {
	m.readFlags()
	m.loadNote()
	m.applyFlags()
	m.loadTags()
	m.askMissing()
	return m.cache.input
}

// Version is the note's updated_at when it was read, so the save is refused if it changed while the form was open.
func (m *mill) Version() time.Time {
	return m.cache.version
}

// Methods

func (m *mill) readFlags() {
	var errs [4]error
	m.cache.content, errs[0] = m.flags.GetString("content")
	m.cache.addTags, errs[1] = m.flags.GetStringSlice("tag")
	m.cache.removeTags, errs[2] = m.flags.GetStringSlice("untag")
	m.cache.tagFromList, errs[3] = m.flags.GetBool("tag-from-list")
	m.SetErr(errors.Join(errs[:]...))
}

func (m *mill) loadNote() {
	if m.HasErr() {
		return
	}

	process := readitem.Process(m.ctx, m.db, domain.NOTE, m.id)
	process.Run()
	current := process.GetResult()
	m.cache.input = editnote.Input{ID: current.Item.ID, Content: current.Content, Tags: current.Tags}
	// A note never edited has no updated_at: its version is created_at, as patchnote compares it.
	m.cache.version = current.Item.UpdatedAt
	if m.cache.version.IsZero() {
		m.cache.version = current.Item.CreatedAt
	}
	m.SetErr(process.GetErr())
}

func (m *mill) applyFlags() {
	if m.HasErr() {
		return
	}

	if m.flags.Changed("content") {
		m.cache.input.Content = m.cache.content
	}
	m.cache.input.Tags = mergeTags(hashtag.Parse(m.cache.input.Content), m.cache.addTags, m.cache.removeTags)
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

// hasChanges reports whether a flag already says what to change, in which case the content is not asked.
func (m *mill) hasChanges() bool {
	return m.flags.Changed("content") || len(m.cache.addTags) > 0 || len(m.cache.removeTags) > 0
}

// Helpers

// mergeTags adds then removes tags, case-insensitively like the tags index; current order and spelling are kept.
func mergeTags(current, add, remove []string) []string {
	has := func(values []string, value string) bool {
		return slices.ContainsFunc(values, func(v string) bool { return strings.EqualFold(strings.TrimSpace(v), strings.TrimSpace(value)) })
	}

	var result []string
	for _, tag := range append(slices.Clone(current), add...) {
		if strings.TrimSpace(tag) != "" && !has(result, tag) && !has(remove, tag) {
			result = append(result, tag)
		}
	}
	return result
}
