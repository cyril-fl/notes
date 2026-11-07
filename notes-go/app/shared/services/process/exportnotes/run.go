package exportnotes

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/wikilink"
)

var ErrFolderNotFound = errors.New("folder not found")

// File is one entry of the export, its Path relative to the export directory: a folder (Dir) or a note.
type File struct {
	Path    string
	Dir     bool
	Content string
}

// item is a live item of the exported tree.
type item struct {
	ID       string      `bun:"id"`
	ParentID string      `bun:"parent_id"`
	Kind     domain.Kind `bun:"kind"`
	Title    string      `bun:"title"`
	Content  string      `bun:"content"`
}

type run struct {
	plumprimitive.Process
	cache

	ctx      context.Context
	db       bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	folderID string
}

type cache struct {
	children map[string][]item
	files    []File
}

// Process exports the live content of a folder ("" = everything) as markdown files: a folder is a
// directory, a note is <title>.md with its text, links to notes written back as [[Title]].
func Process(ctx context.Context, db bun.IDB, folderID string) *run {
	return &run{
		ctx:      ctx,
		db:       db,
		folderID: folderID,
	}
}

func (p *run) Run() {
	p.checkFolder()
	p.loadTree()
	p.makeFiles(p.folderID, "")
}

// Getters & Setters

// GetResult returns the files, each folder before what it holds.
func (p *run) GetResult() []File {
	return p.cache.files
}

// Methods

func (p *run) checkFolder() {
	if p.folderID == "" {
		return
	}
	exists, err := p.db.NewSelect().Model((*domain.Item)(nil)).Where("id = ? AND kind = ?", p.folderID, domain.FOLDER).Exists(p.ctx)
	if err == nil && !exists {
		err = fmt.Errorf("%w: %q", ErrFolderNotFound, p.folderID)
	}
	p.SetErr(err)
}

func (p *run) loadTree() {
	if p.HasErr() {
		return
	}

	var items []item
	err := p.db.NewRaw(`
		WITH RECURSIVE tree AS (
			SELECT id FROM items WHERE deleted_at IS NULL AND (CASE WHEN ? = '' THEN parent_id IS NULL ELSE parent_id::text = ? END)
			UNION ALL SELECT child.id FROM items AS child JOIN tree ON child.parent_id = tree.id WHERE child.deleted_at IS NULL
		)
		SELECT i.id, coalesce(i.parent_id::text, '') AS parent_id, i.kind, coalesce(i.title, '') AS title, coalesce(mn.content, '') AS content
		FROM tree
		JOIN items AS i ON i.id = tree.id
		LEFT JOIN metadata_notes AS mn ON mn.item_id = i.id
		ORDER BY i.kind, lower(i.title), i.created_at`,
		p.folderID, p.folderID).
		Scan(p.ctx, &items)
	if err != nil {
		p.SetErr(err)
		return
	}

	p.cache.children = map[string][]item{}
	for _, it := range items {
		p.cache.children[it.ParentID] = append(p.cache.children[it.ParentID], it)
	}
}

// makeFiles names the children of a folder: names unique whatever their case (macOS and Windows disks ignore it).
func (p *run) makeFiles(parentID, dir string) {
	if p.HasErr() {
		return
	}

	taken := map[string]bool{}
	for _, it := range p.cache.children[parentID] {
		ext := ""
		if it.Kind == domain.NOTE {
			ext = ".md"
		}
		name := uniqueName(fileName(it.Title), ext, taken)
		file := File{Path: path.Join(dir, name), Dir: it.Kind == domain.FOLDER, Content: wikilink.Unresolve(it.Content)}
		p.cache.files = append(p.cache.files, file)
		if file.Dir {
			p.makeFiles(it.ID, file.Path)
		}
	}
}

// Helpers

// fileName keeps a title as it is, but for what a file name cannot hold.
func fileName(title string) string {
	name := strings.TrimLeft(strings.TrimSpace(strings.NewReplacer("/", "-", "\\", "-", "\x00", "").Replace(title)), ".")
	if name == "" {
		return "untitled"
	}
	return name
}

func uniqueName(name, ext string, taken map[string]bool) string {
	candidate := name + ext
	for n := 2; taken[strings.ToLower(candidate)]; n++ {
		candidate = fmt.Sprintf("%s (%d)%s", name, n, ext)
	}
	taken[strings.ToLower(candidate)] = true
	return candidate
}
