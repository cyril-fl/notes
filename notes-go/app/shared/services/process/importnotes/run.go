package importnotes

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/notetitle"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createfile"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createfolder"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/exportnotes"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/silo"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

// imagePattern finds the markdown images ![alt](target "title") and ![alt](<target with spaces>): the target is
// the second group, or the third.
var imagePattern = regexp.MustCompile(`(!\[[^\]]*\]\()(?:<([^>]+)>|([^)\s]+))`)

// embedPattern finds Obsidian's ![[image.png]] and ![[image.png|300]] (or |300x200): the name is the first group,
// the width the second.
var embedPattern = regexp.MustCompile(`!\[\[([^\]|]+)(?:\|(\d*)[^\]]*)?\]\]`)

// sizeTitle is the title of the markdown that gives an image of an Obsidian width in pixels the editor's closest
// size, none above 600 or for no width: the image keeps its own then.
func sizeTitle(width string) string {
	n, err := strconv.Atoi(width)
	switch {
	case err != nil:
		return ""
	case n <= 300:
		return ` "small"`
	case n <= 600:
		return ` "medium"`
	}
	return ""
}

// Input is what to import: every directory and .md file of Dir, into the folder ParentID ("" = the root).
type Input struct {
	Dir      fs.FS
	ParentID string
}

// Result counts what was created.
type Result struct {
	Folders int `json:"folders"`
	Notes   int `json:"notes"`
}

type run struct {
	plumprimitive.Process
	cache

	ctx   context.Context
	db    bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	store silo.Store
	input Input
}

type cache struct {
	folders map[string]string      // directory path -> folder id
	images  map[string]domain.File // file path -> the image made of it
	names   map[string][]string    // file name -> its paths in Dir, read at the first ![[name]]
	result  Result
}

// Process imports markdown files as notes and their directories as folders, all or nothing: a folder
// name already taken stops it. Titles, #tags and [[links]] come from the text, as for any note; the images
// a note cites by a relative path or by ![[name]] become files of the account, exportnotes.FilesDir no folder.
func Process(ctx context.Context, db bun.IDB, store silo.Store, input Input) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		store: store,
		input: input,
	}
}

// Run removes the bytes of the images made when it fails: the rolled-back rows no longer cite them.
func (p *run) Run() {
	err := p.db.RunInTx(p.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		p.cache = cache{folders: map[string]string{".": p.input.ParentID}, images: map[string]domain.File{}}
		return fs.WalkDir(p.input.Dir, ".", func(name string, entry fs.DirEntry, err error) error {
			if err != nil || name == "." {
				return err
			}
			return p.importEntry(ctx, tx, name, entry)
		})
	})
	if err != nil {
		// A client gone has canceled p.ctx: the cleanup must still reach Silo.
		cleanup := context.WithoutCancel(p.ctx)
		for _, file := range p.cache.images {
			_ = p.store.Remove(cleanup, file.Key())
		}
	}
	p.SetErr(err)
}

// Getters & Setters

func (p *run) GetResult() Result {
	return p.cache.result
}

// Methods

// importEntry skips hidden entries (.git, .obsidian…, the __MACOSX of a Finder zip), files other than .md and blank notes.
func (p *run) importEntry(ctx context.Context, tx bun.Tx, name string, entry fs.DirEntry) error {
	hidden := strings.HasPrefix(entry.Name(), ".") || entry.Name() == "__MACOSX"
	if entry.IsDir() {
		if hidden || name == exportnotes.FilesDir {
			return fs.SkipDir
		}
		return p.importFolder(ctx, tx, name, entry.Name())
	}
	if hidden || !strings.EqualFold(path.Ext(name), ".md") {
		return nil
	}

	data, err := fs.ReadFile(p.input.Dir, name)
	if err != nil || strings.TrimSpace(string(data)) == "" {
		return err
	}
	content, err := p.importImages(ctx, tx, name, string(data))
	if err != nil {
		return err
	}
	note := createnote.Process(ctx, tx, createnote.Input{Content: withTitle(content, name), ParentID: p.cache.folders[path.Dir(name)]})
	note.Run()
	if note.HasErr() {
		return fmt.Errorf("%s: %w", name, note.GetErr())
	}
	p.cache.result.Notes++
	return nil
}

// importImages makes a file of each image the note at name cites by a relative path or by ![[name]], once per
// image, and cites it by its address instead. A missing file or one that is not an image keeps its link as written.
func (p *run) importImages(ctx context.Context, tx bun.Tx, name, content string) (string, error) {
	var failed error
	content = imagePattern.ReplaceAllStringFunc(content, func(link string) string {
		parts := imagePattern.FindStringSubmatch(link)
		target, err := url.PathUnescape(parts[2] + parts[3])
		if err != nil || failed != nil || strings.Contains(target, ":") || strings.HasPrefix(target, "/") {
			return link
		}
		file, ok, err := p.importImage(ctx, tx, path.Join(path.Dir(name), target))
		if !ok {
			failed = err
			return link
		}
		return parts[1] + domain.FileURL(file.ID)
	})
	content = embedPattern.ReplaceAllStringFunc(content, func(link string) string {
		embed := embedPattern.FindStringSubmatch(link)
		target, ok := p.findEmbed(path.Dir(name), embed[1])
		if !ok || failed != nil {
			return link
		}
		file, ok, err := p.importImage(ctx, tx, target)
		if !ok {
			failed = err
			return link
		}
		return "![](" + domain.FileURL(file.ID) + sizeTitle(embed[2]) + ")"
	})
	return content, failed
}

// importImage makes a file of the image at target, once. ok is false for a missing file or one that is not an
// image, err is set only when making it failed otherwise.
func (p *run) importImage(ctx context.Context, tx bun.Tx, target string) (domain.File, bool, error) {
	if file, ok := p.cache.images[target]; ok {
		return file, true, nil
	}
	data, err := fs.ReadFile(p.input.Dir, target)
	if err != nil {
		return domain.File{}, false, nil
	}
	image := createfile.Process(ctx, tx, p.store, createfile.Input{Name: path.Base(target), Data: data})
	image.Run()
	if image.HasErr() {
		if errors.Is(image.GetErr(), createfile.ErrNotImage) {
			return domain.File{}, false, nil
		}
		return domain.File{}, false, fmt.Errorf("%s: %w", target, image.GetErr())
	}
	p.cache.images[target] = image.GetResult()
	return image.GetResult(), true, nil
}

// findEmbed finds the file a ![[target]] of a note in dir cites: Obsidian writes its name, or the end of its path
// when the name is not unique. The one in the note's directory wins, then the first in Dir's order.
func (p *run) findEmbed(dir, target string) (string, bool) {
	if p.cache.names == nil {
		p.cache.names = map[string][]string{}
		_ = fs.WalkDir(p.input.Dir, ".", func(name string, entry fs.DirEntry, err error) error {
			if err == nil && !entry.IsDir() {
				p.cache.names[entry.Name()] = append(p.cache.names[entry.Name()], name)
			}
			return nil
		})
	}
	target = strings.TrimSpace(target)
	found := ""
	for _, name := range p.cache.names[path.Base(target)] {
		if name != target && !strings.HasSuffix(name, "/"+target) {
			continue
		}
		if path.Dir(name) == dir {
			return name, true
		}
		if found == "" {
			found = name
		}
	}
	return found, found != ""
}

// withTitle writes the file name as a heading on a text without one (Obsidian titles a note by its file),
// unless its first words already give that name: an untitled note exported by notes comes back as it was.
func withTitle(content, name string) string {
	title := strings.TrimSuffix(path.Base(name), path.Ext(name))
	if notetitle.Heading(content) != "" || notetitle.From(content) == title {
		return content
	}
	return "# " + title + "\n\n" + content
}

func (p *run) importFolder(ctx context.Context, tx bun.Tx, name, title string) error {
	folder := createfolder.Process(ctx, tx, createfolder.Input{Title: title, ParentID: p.cache.folders[path.Dir(name)]})
	folder.Run()
	if folder.HasErr() {
		return fmt.Errorf("%s: %w", name, folder.GetErr())
	}
	p.cache.folders[name] = folder.GetResult().ID
	p.cache.result.Folders++
	return nil
}
