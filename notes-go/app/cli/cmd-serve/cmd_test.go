package cmdserve

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestEvery(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var runs atomic.Int32
	done := make(chan struct{})
	go func() {
		every(ctx, time.Millisecond, func() { runs.Add(1) })
		close(done)
	}()

	deadline := time.After(5 * time.Second)
	for runs.Load() < 2 {
		select {
		case <-deadline:
			t.Fatal("the task did not run twice")
		case <-time.After(time.Millisecond):
		}
	}

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("every did not stop with its context")
	}
}

func TestWithWeb(t *testing.T) {
	dir := t.TempDir()
	for name, body := range map[string]string{"200.html": "app", "app.js": "js"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("api")) })
	handler := withWeb(dir, api)

	for path, want := range map[string]string{
		"/api/me":   "api",
		"/app.js":   "js",
		"/notes/42": "app",
	} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if got := strings.TrimSpace(w.Body.String()); got != want {
			t.Errorf("%s: body = %q, want %q", path, got, want)
		}
	}
}

func TestWithProxy(t *testing.T) {
	handler := withProxy(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(r.RemoteAddr)) }))

	for name, test := range map[string]struct {
		forwarded []string
		want      string
	}{
		"no header":          {nil, "192.0.2.1:1234"},
		"one address":        {[]string{"100.64.0.2"}, "100.64.0.2:0"},
		"forged first entry": {[]string{"1.2.3.4, 100.64.0.2"}, "100.64.0.2:0"},
		"repeated header":    {[]string{"1.2.3.4", "100.64.0.3"}, "100.64.0.3:0"},
		"ipv6":               {[]string{"fd7a:115c::1"}, "[fd7a:115c::1]:0"},
		"empty":              {[]string{""}, "192.0.2.1:1234"},
	} {
		r := httptest.NewRequest(http.MethodPost, "/api/login", nil)
		for _, value := range test.forwarded {
			r.Header.Add("X-Forwarded-For", value)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if got := w.Body.String(); got != test.want {
			t.Errorf("%s: RemoteAddr = %q, want %q", name, got, test.want)
		}
	}
}
