package domain

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
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

// CompareFold orders two titles, paths or tags whatever their case, as lower() did in SQL.
func CompareFold(a, b string) int {
	return strings.Compare(strings.ToLower(a), strings.ToLower(b))
}
