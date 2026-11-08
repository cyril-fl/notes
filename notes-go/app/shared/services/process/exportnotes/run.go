package exportnotes

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/wikilink"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

var ErrFolderNotFound = errors.New("folder not found")

// FilesDir holds the images of the export, at its root.
const FilesDir = "_files"

// File is one entry of the export, its Path relative to the export directory: a folder (Dir), a note, or an
// image whose bytes the writer reads from Silo at Key.
type File struct {
	Path    string
	Dir     bool
	Content string
	Key     string
}

// item is a live item of the exported tree.
type item struct {
	ID       string      `bun:"id"`
	ParentID string      `bun:"parent_id"`
	Kind     domain.Kind `bun:"kind"`
	Title    secret.Text `bun:"title"`
	Content  secret.Text `bun:"content"`
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
	images   map[string]domain.File // the files the exported notes cite, by id
	files    []File
}

// Process exports the live content of a folder ("" = everything) as markdown files: a folder is a
// directory, a note is <title>.md with its text, links to notes written back as [[Title]], images
// copied into FilesDir and cited by a relative path.
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
	p.loadImages()
	p.makeFiles(p.folderID, "")
	p.addImages()
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
	exists, err := p.db.NewSelect().Model((*domain.Item)(nil)).Where("id = ? AND kind = ? AND user_id = ?", p.folderID, domain.FOLDER, domain.UserID(p.ctx)).Exists(p.ctx)
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
			SELECT id FROM items WHERE deleted_at IS NULL AND user_id = ? AND (CASE WHEN ? = '' THEN parent_id IS NULL ELSE parent_id::text = ? END)
			UNION ALL SELECT child.id FROM items AS child JOIN tree ON child.parent_id = tree.id WHERE child.deleted_at IS NULL
		)
		SELECT i.id, coalesce(i.parent_id::text, '') AS parent_id, i.kind, i.title, mn.content
		FROM tree
		JOIN items AS i ON i.id = tree.id
		LEFT JOIN metadata_notes AS mn ON mn.item_id = i.id
		ORDER BY i.kind, i.created_at`,
		domain.UserID(p.ctx), p.folderID, p.folderID).
		Scan(p.ctx, &items)
	if err != nil {
		p.SetErr(err)
		return
	}

	// Titles are sealed: ordered here, the oldest first among the same title, as uniqueName numbers them.
	domain.SortFold(items, func(i item) string { return string(i.Title) }, func(a, b item) int { return strings.Compare(string(a.Kind), string(b.Kind)) })
	p.cache.children = map[string][]item{}
	for _, it := range items {
		p.cache.children[it.ParentID] = append(p.cache.children[it.ParentID], it)
	}
}

// loadImages keeps the files of the account only: a note may cite the address of another account's.
func (p *run) loadImages() {
	if p.HasErr() {
		return
	}

	var ids []string
	for _, items := range p.cache.children {
		for _, it := range items {
			for _, match := range domain.FilePattern.FindAllStringSubmatch(string(it.Content), -1) {
				ids = append(ids, match[1])
			}
		}
	}
	p.cache.images = map[string]domain.File{}
	if len(ids) == 0 {
		return
	}

	var files []domain.File
	err := p.db.NewSelect().Model(&files).Where("id IN (?) AND user_id = ?", bun.List(ids), domain.UserID(p.ctx)).Scan(p.ctx)
	for _, file := range files {
		p.cache.images[file.ID] = file
	}
	p.SetErr(err)
}

// makeFiles names the children of a folder: names unique whatever their case (macOS and Windows disks ignore it).
func (p *run) makeFiles(parentID, dir string) {
	if p.HasErr() {
		return
	}

	// A folder named like FilesDir at the root becomes "_files (2)".
	taken := map[string]bool{strings.ToLower(FilesDir): dir == "" && len(p.cache.images) > 0}
	for _, it := range p.cache.children[parentID] {
		ext := ""
		if it.Kind == domain.NOTE {
			ext = ".md"
		}
		name := uniqueName(fileName(string(it.Title)), ext, taken)
		file := File{Path: path.Join(dir, name), Dir: it.Kind == domain.FOLDER, Content: p.citeImages(wikilink.Unresolve(string(it.Content)), dir)}
		p.cache.files = append(p.cache.files, file)
		if file.Dir {
			p.makeFiles(it.ID, file.Path)
		}
	}
}

// citeImages points the images of a note in dir to their copy in FilesDir.
func (p *run) citeImages(content, dir string) string {
	up := ""
	if dir != "" {
		up = strings.Repeat("../", strings.Count(dir, "/")+1)
	}
	return domain.FilePattern.ReplaceAllStringFunc(content, func(url string) string {
		file, ok := p.cache.images[domain.FilePattern.FindStringSubmatch(url)[1]]
		if !ok {
			return url
		}
		return up + imagePath(file)
	})
}

// addImages adds FilesDir and the images, sorted: the export of the same notes gives the same files.
func (p *run) addImages() {
	if p.HasErr() || len(p.cache.images) == 0 {
		return
	}

	p.cache.files = append(p.cache.files, File{Path: FilesDir, Dir: true})
	ids := slices.Sorted(maps.Keys(p.cache.images))
	for _, id := range ids {
		file := p.cache.images[id]
		p.cache.files = append(p.cache.files, File{Path: imagePath(file), Key: file.Key()})
	}
}

// Helpers

// imagePath names an image by its id: two pasted images are both "image.png".
func imagePath(file domain.File) string {
	return FilesDir + "/" + file.ID + domain.FileTypes[file.ContentType]
}

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
