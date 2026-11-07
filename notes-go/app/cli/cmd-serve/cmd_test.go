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
