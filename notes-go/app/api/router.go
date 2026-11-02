// Package api is the HTTP API: one folder per route, each calling the same processes as the CLI.
package api

import (
	"net/http"

	"github.com/uptrace/bun"
)

// New gives the API, reachable by web pages from origin only (none when empty).
func New(db bun.IDB, origin string) http.Handler {
	mux := http.NewServeMux()

	return mux
}
