package login

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestThrottle(t *testing.T) {
	now := time.Now()
	th := &throttle{now: func() time.Time { return now }, byKey: map[string]strikes{}}

	want := []time.Duration{0, 0, 0, 30 * time.Second, time.Minute, 5 * time.Minute, 15 * time.Minute, time.Hour, time.Hour}
	for i, wait := range want {
		th.fail("a")
		if got := th.wait("a"); got != wait {
			t.Errorf("after %d failures: wait = %v, want %v", i+1, got, wait)
		}
		now = now.Add(wait)
		if got := th.wait("a"); got != 0 {
			t.Errorf("after %d failures and the wait: wait = %v, want 0", i+1, got)
		}
	}

	if got := th.wait("b"); got != 0 {
		t.Errorf("other key: wait = %v, want 0", got)
	}

	th.fail("a")
	now = now.Add(10 * time.Minute)
	w := httptest.NewRecorder()
	if !th.blocked(w, "a") || w.Code != http.StatusTooManyRequests || w.Header().Get("Retry-After") != "3000" {
		t.Errorf("blocked: status = %d, Retry-After = %q, want 429 and 3000", w.Code, w.Header().Get("Retry-After"))
	}

	th.reset("a")
	th.fail("a")
	if got := th.wait("a"); got != 0 {
		t.Errorf("first failure after a reset: wait = %v, want 0", got)
	}
}

func TestThrottleAtOnce(t *testing.T) {
	th := &throttle{now: time.Now, byKey: map[string]strikes{}}

	var wg sync.WaitGroup
	var mu sync.Mutex
	passed := 0
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !th.blocked(httptest.NewRecorder(), "a") {
				mu.Lock()
				passed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if passed != 4 {
		t.Errorf("tries sent at once: %d passed, want 4", passed)
	}

	th.forgive("a")
	if got := th.byKey["a"].count; got != 3 {
		t.Errorf("after forgive: count = %d, want 3", got)
	}
}
