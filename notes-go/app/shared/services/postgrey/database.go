package postgrey

import "github.com/uptrace/bun"

var db *bun.DB

// GetDb returns the shared connection, nil before SetDb.
func GetDb() *bun.DB {
	return db
}

// SetDb sets the shared connection.
func SetDb(connection *bun.DB) {
	db = connection
}
