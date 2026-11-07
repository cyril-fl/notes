package domain

import (
	"errors"

	"github.com/uptrace/bun/driver/pgdriver"
)

// FolderNameIndex keeps two live folders of one folder (or of an account's root) from sharing a name, whatever its case.
const FolderNameIndex = "ux_folders_user_parent_title_hash"

// uniqueViolation is the Postgres SQLSTATE raised by FolderNameIndex.
const uniqueViolation = "23505"

var ErrFolderExists = errors.New("a folder with this name already exists here")

// FolderNameTaken tells whether err is FolderNameIndex refusing a folder: created, renamed, moved or restored
// next to one of the same name.
func FolderNameTaken(err error) bool {
	var pgErr pgdriver.Error
	return errors.As(err, &pgErr) && pgErr.Field('C') == uniqueViolation && pgErr.Field('n') == FolderNameIndex
}
