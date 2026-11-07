package domain

import "github.com/uptrace/bun"

// RegisterModels must run before any query on MetadataNote.Tags (bun m2m requirement).
func RegisterModels(db *bun.DB) {
	db.RegisterModel((*NoteTag)(nil))
}

// PathsCTE gives every live item of an account its path from its root (/Projects/Go/Ideas): append a query on
// "paths", the account's id as first argument. It walks down from the roots, so a trashed folder hides its whole subtree.
const PathsCTE = `WITH RECURSIVE paths AS (
	SELECT id, '/' || coalesce(title, 'untitled') AS path FROM items WHERE parent_id IS NULL AND deleted_at IS NULL AND user_id = ?
	UNION ALL SELECT child.id, paths.path || '/' || coalesce(child.title, 'untitled')
	FROM items AS child JOIN paths ON child.parent_id = paths.id WHERE child.deleted_at IS NULL
) `
