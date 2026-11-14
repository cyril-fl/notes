package domain

import (
	"bytes"
	"context"
	"slices"
	"sync"

	"github.com/uptrace/bun"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// RegisterModels must run before any query on MetadataNote.Tags (bun m2m requirement).
func RegisterModels(db *bun.DB) {
	db.RegisterModel((*NoteTag)(nil))
}

// LiveCTE gives every live item of an account, as "live" (id): append a query on it, the account's id as first argument.
// It walks down from the roots, so a trashed folder hides its whole subtree. Paths gives their paths.
const LiveCTE = `WITH RECURSIVE live AS (
	SELECT id FROM items WHERE parent_id IS NULL AND deleted_at IS NULL AND user_id = ?
	UNION ALL SELECT child.id FROM items AS child JOIN live ON child.parent_id = live.id WHERE child.deleted_at IS NULL
) `

// Paths gives every live item of the account of ctx its path from its root (/Projects/Go/Ideas), by id: a trashed
// folder hides its whole subtree, so an id missing from it is not live. Titles are sealed: paths are made here, not in SQL.
// Cached per account while users.items_version stays the same (one query, not 75 ms of decryption at 50,000 items):
// the map is shared, do not modify it.
func Paths(ctx context.Context, db bun.IDB) (map[string]string, error) {
	userID := UserID(ctx)
	// The version before the items: a write committed between the two only makes the next call rebuild.
	var version int64
	if err := db.NewSelect().Model((*User)(nil)).Column("items_version").Where("id = ?", userID).Scan(ctx, &version); err != nil {
		return nil, err
	}
	if cached, ok := pathsCache.Load(userID); ok && cached.(cachedPaths).version == version {
		return cached.(cachedPaths).paths, nil
	}

	var items []Item
	err := db.NewSelect().Model(&items).Column("id", "parent_id", "title").Where("user_id = ?", userID).Scan(ctx)
	if err != nil {
		return nil, err
	}

	children := map[string][]Item{}
	for _, item := range items {
		parent := ""
		if item.ParentID != nil {
			parent = *item.ParentID
		}
		children[parent] = append(children[parent], item)
	}
	paths := map[string]string{}
	var walk func(parent, prefix string)
	walk = func(parent, prefix string) {
		for _, item := range children[parent] {
			title := string(item.Title)
			if title == "" {
				title = "untitled"
			}
			paths[item.ID] = prefix + "/" + title
			walk(item.ID, paths[item.ID])
		}
	}
	walk("", "")
	pathsCache.Store(userID, cachedPaths{version: version, paths: paths})
	return paths, nil
}

// pathsCache keeps the last Paths of each account, by account id. Titles in clear: in memory only, never written.
var pathsCache sync.Map

type cachedPaths struct {
	version int64
	paths   map[string]string
}

// collators are pooled: a collate.Collator is not safe for concurrent use.
// The root order is the same for fr and en (CLDR has no fr-FR rules), so the account's locale is not needed.
var collators = sync.Pool{New: func() any { return collate.New(language.Und, collate.IgnoreCase) }}

// SortFold sorts items by first, then by key whatever its case and with accented letters next to their base letter
// (Été between abricot and zoo), as the collation of Postgres did. first may be nil; the sort is stable.
// Each key is collated once, not on every comparison: 50,000 paths sort in tens of ms, not hundreds.
func SortFold[T any](items []T, key func(T) string, first func(a, b T) int) {
	c := collators.Get().(*collate.Collator)
	defer collators.Put(c)
	// Indexes sorted, not the items: a stable sort moves big items in O(n log² n), 150 ms for 50,000 notes.
	type keyed struct {
		index int
		key   []byte
	}
	var buf collate.Buffer
	all := make([]keyed, len(items))
	for i, item := range items {
		all[i] = keyed{i, c.KeyFromString(&buf, key(item))}
	}
	slices.SortFunc(all, func(a, b keyed) int {
		if first != nil {
			if order := first(items[a.index], items[b.index]); order != 0 {
				return order
			}
		}
		if order := bytes.Compare(a.key, b.key); order != 0 {
			return order
		}
		return a.index - b.index // stable
	})
	sorted := make([]T, len(items))
	for i, k := range all {
		sorted[i] = items[k.index]
	}
	copy(items, sorted)
}
