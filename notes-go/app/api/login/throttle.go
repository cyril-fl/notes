package login

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"main.go/app/api/services/respond"
)

// waits is how long to wait after 1, 2, 3... wrong passwords in a row, like an iPhone: a password cannot be guessed by trying them all.
var waits = []time.Duration{0, 0, 0, 30 * time.Second, time.Minute, 5 * time.Minute, 15 * time.Minute, time.Hour}

// throttle counts the wrong passwords in a row, by address for the login and by account for a new password.
// shortcut: kept in memory, a restart of the API forgets it; enough behind Tailscale, store it if the API ever faces the internet.
type throttle struct {
	sync.Mutex
	now   func() time.Time
	byKey map[string]strikes
}

type strikes struct {
	count int
	last  time.Time
}

// failed is shared by the two routes that check a password.
var failed = &throttle{now: time.Now, byKey: map[string]strikes{}}

// wait is how long key has left to wait before its next try.
func (t *throttle) wait(key string) time.Duration {
	t.Lock()
	defer t.Unlock()
	s := t.byKey[key]
	if s.count == 0 {
		return 0
	}
	return max(waits[min(s.count, len(waits))-1]-t.now().Sub(s.last), 0)
}

func (t *throttle) fail(key string) {
	t.Lock()
	defer t.Unlock()
	t.byKey[key] = strikes{count: t.byKey[key].count + 1, last: t.now()}
}

func (t *throttle) reset(key string) {
	t.Lock()
	defer t.Unlock()
	delete(t.byKey, key)
}

// blocked answers 429 with Retry-After while key waits.
func (t *throttle) blocked(w http.ResponseWriter, key string) bool {
	wait := t.wait(key)
	if wait == 0 {
		return false
	}
	w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(wait.Seconds()))))
	respond.JSON(w, http.StatusTooManyRequests, respond.Failure{Error: "too many wrong passwords, try again later"})
	return true
}
