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
// shortcut: loads and decrypts every item of the account on each call, 7 ms for 5,000 notes (#23); cache it per account past ~50,000.
func Paths(ctx context.Context, db bun.IDB) (map[string]string, error) {
	var items []Item
	err := db.NewSelect().Model(&items).Column("id", "parent_id", "title").Where("user_id = ?", UserID(ctx)).Scan(ctx)
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
	return paths, nil
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
	type keyed struct {
		item T
		key  []byte
	}
	var buf collate.Buffer
	all := make([]keyed, len(items))
	for i, item := range items {
		all[i] = keyed{item, c.KeyFromString(&buf, key(item))}
	}
	slices.SortStableFunc(all, func(a, b keyed) int {
		if first != nil {
			if order := first(a.item, b.item); order != 0 {
				return order
			}
		}
		return bytes.Compare(a.key, b.key)
	})
	for i := range all {
		items[i] = all[i].item
	}
}
