// Package auth lets through the requests of a logged-in account only, and tells the routes which one.
package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createsession"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readsession"
	"github.com/uptrace/bun"
)

// Cookie holds the session token of the web app: HttpOnly, page scripts never see it.
const Cookie = "notes_session"

// Public routes answer without a session.
var public = map[string]bool{"POST /api/login": true}

// New runs next with the account of the request's session in its context: the cookie, or a
// "Authorization: Bearer <token>" header for scripts. Without a valid one, 401.
func New(db bun.IDB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if public[r.Method+" "+r.URL.Path] || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		token := Token(r)
		session := readsession.Process(r.Context(), db, token)
		session.Run()
		if session.HasErr() {
			respond.Error(w, session.GetErr())
			return
		}
		// readsession started the session's lifetime again: the cookie follows.
		SetCookie(w, token, time.Now().Add(createsession.Lifetime))
		next.ServeHTTP(w, r.WithContext(domain.WithUser(r.Context(), session.GetResult().ID)))
	})
}

// Token is the request's session token, "" without one.
func Token(r *http.Request) string {
	if bearer, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
		return bearer
	}
	if cookie, err := r.Cookie(Cookie); err == nil {
		return cookie.Value
	}
	return ""
}

// SetCookie gives the browser the session token until expires; a zero time deletes it.
// Secure works on http://localhost too: browsers count it as a secure context.
func SetCookie(w http.ResponseWriter, token string, expires time.Time) {
	cookie := &http.Cookie{Name: Cookie, Value: token, Path: "/", HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode, Expires: expires}
	if expires.IsZero() {
		cookie.MaxAge = -1
	}
	http.SetCookie(w, cookie)
}
