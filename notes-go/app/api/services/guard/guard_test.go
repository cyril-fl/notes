package guard

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

const web = "http://localhost:3001"

func TestNew(t *testing.T) {
	tests := map[string]struct {
		allowed string
		method  string
		host    string
		origin  string
		status  int
		cors    bool
	}{
		"no origin header":      {web, http.MethodGet, "localhost:8080", "", http.StatusOK, false},
		"allowed origin":        {web, http.MethodPost, "localhost:8080", web, http.StatusOK, true},
		"other origin":          {web, http.MethodGet, "localhost:8080", "https://evil.example", http.StatusForbidden, false},
		"other origin, delete":  {web, http.MethodDelete, "127.0.0.1:8080", "https://evil.example", http.StatusForbidden, false},
		"same host, other port": {web, http.MethodGet, "localhost:8080", "http://localhost:9999", http.StatusForbidden, false},
		"none configured":       {"", http.MethodGet, "localhost:8080", web, http.StatusForbidden, false},
		"none configured, curl": {"", http.MethodGet, "127.0.0.1:8080", "", http.StatusOK, false},
		"preflight":             {web, http.MethodOptions, "localhost:8080", web, http.StatusNoContent, true},
		"preflight, other":      {web, http.MethodOptions, "localhost:8080", "https://evil.example", http.StatusForbidden, false},
		"rebinding host":        {web, http.MethodGet, "evil.example", "", http.StatusForbidden, false},
		"rebinding with port":   {web, http.MethodGet, "evil.example:8080", "", http.StatusForbidden, false},
		"ipv6 loopback":         {web, http.MethodGet, "[::1]:8080", "", http.StatusOK, false},
		"ipv6 without port":     {web, http.MethodGet, "[::1]", "", http.StatusOK, false},
		"host of the origin":    {"http://notes.lan:3001", http.MethodGet, "notes.lan:8080", "", http.StatusOK, false},
	}

	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	for name, test := range tests {
		r := httptest.NewRequest(test.method, "/", nil)
		r.Host = test.host
		if test.origin != "" {
			r.Header.Set("Origin", test.origin)
		}

		w := httptest.NewRecorder()
		New(test.allowed, ok).ServeHTTP(w, r)
		if w.Code != test.status {
			t.Errorf("%s: status = %d, want %d", name, w.Code, test.status)
		}
		if got := w.Header().Get("Access-Control-Allow-Origin") != ""; got != test.cors {
			t.Errorf("%s: cors header = %v, want %v", name, got, test.cors)
		}
	}
}

func TestCheckOrigin(t *testing.T) {
	for origin, want := range map[string]error{
		"":                      nil,
		web:                     nil,
		"https://notes.example": nil,
		web + "/":               ErrInvalidOrigin,
		"localhost:3001":        ErrInvalidOrigin,
		"notes.example":         ErrInvalidOrigin,
	} {
		if err := CheckOrigin(origin); !errors.Is(err, want) {
			t.Errorf("CheckOrigin(%q) = %v, want %v", origin, err, want)
		}
	}
}
