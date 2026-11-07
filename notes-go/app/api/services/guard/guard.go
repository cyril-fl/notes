// Package guard keeps the web pages of other sites away from the API: they could use the session cookie.
package guard

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"

	"main.go/app/api/services/respond"
)

var ErrInvalidOrigin = errors.New("origin must look like http://localhost:3000, with no path")

// CheckOrigin refuses what could never match a browser Origin header: a trailing slash would silently block the web app.
func CheckOrigin(origin string) error {
	if origin == "" {
		return nil
	}

	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.Path != "" {
		return ErrInvalidOrigin
	}
	return nil
}

// New lets through requests from origin, and those with no Origin at all (curl, scripts), which no web page can send
// to read or change anything. The Host check stops a domain pointed at 127.0.0.1 (DNS rebinding): its pages are
// same-origin for the browser, so they send no Origin on a GET.
func New(origin string, next http.Handler) http.Handler {
	hosts := map[string]bool{"localhost": true, "127.0.0.1": true, "::1": true}
	if parsed, err := url.Parse(origin); err == nil && parsed.Hostname() != "" {
		hosts[parsed.Hostname()] = true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !hosts[hostname(r.Host)] {
			respond.JSON(w, http.StatusForbidden, respond.Failure{Error: "forbidden host"})
			return
		}

		w.Header().Add("Vary", "Origin")
		from := r.Header.Get("Origin")
		switch {
		case from == "":
			next.ServeHTTP(w, r)
		case from != origin:
			respond.JSON(w, http.StatusForbidden, respond.Failure{Error: "forbidden origin"})
		case r.Method == http.MethodOptions:
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.WriteHeader(http.StatusNoContent)
		default:
			w.Header().Set("Access-Control-Allow-Origin", origin)
			// The web app runs on another port in development: its requests carry the session cookie.
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			// The login and account pages read how long to wait after wrong passwords.
			w.Header().Set("Access-Control-Expose-Headers", "Retry-After")
			next.ServeHTTP(w, r)
		}
	})
}

// hostname drops the port and the brackets of an IPv6 host.
func hostname(host string) string {
	if name, _, err := net.SplitHostPort(host); err == nil {
		return name
	}
	return strings.Trim(host, "[]")
}
