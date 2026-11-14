// Package throttle makes a key wait longer after each try in a row: wrong passwords at the login, accounts created
// from one address at the sign-up.
package throttle

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
)

// waits is how long to wait after 1, 2, 3... tries in a row, like an iPhone: a password cannot be guessed by trying them all.
var waits = []time.Duration{0, 0, 0, 30 * time.Second, time.Minute, 5 * time.Minute, 15 * time.Minute, time.Hour}

// Throttle counts the tries in a row of each key.
// TODO: kept in memory, a restart of the API forgets it; enough behind Tailscale, store it if the API ever faces the internet.
type Throttle struct {
	sync.Mutex
	now   func() time.Time
	byKey map[string]strikes
	// message is the error of the 429.
	message string
}

type strikes struct {
	count int
	last  time.Time
}

// New gives a throttle whose 429 says message.
func New(message string) *Throttle {
	return &Throttle{now: time.Now, byKey: map[string]strikes{}, message: message}
}

// wait is how long key has left to wait before its next try.
func (t *Throttle) wait(key string) time.Duration {
	t.Lock()
	defer t.Unlock()
	return t.waitLocked(key)
}

func (t *Throttle) waitLocked(key string) time.Duration {
	s := t.byKey[key]
	if s.count == 0 {
		return 0
	}
	return max(waits[min(s.count, len(waits))-1]-t.now().Sub(s.last), 0)
}

// fail counts a try for key, t locked.
func (t *Throttle) fail(key string) {
	t.byKey[key] = strikes{count: t.byKey[key].count + 1, last: t.now()}
}

// Forgive takes back the try Blocked counted, when it did not count.
func (t *Throttle) Forgive(key string) {
	t.Lock()
	defer t.Unlock()
	s := t.byKey[key]
	if s.count <= 1 {
		delete(t.byKey, key)
		return
	}
	s.count--
	t.byKey[key] = s
}

func (t *Throttle) Reset(key string) {
	t.Lock()
	defer t.Unlock()
	delete(t.byKey, key)
}

// Blocked answers 429 with Retry-After while key waits, else counts the try before it runs:
// tries sent at once cannot all pass while argon2id runs. The caller resets on a good password, forgives on another error.
func (t *Throttle) Blocked(w http.ResponseWriter, key string) bool {
	t.Lock()
	wait := t.waitLocked(key)
	if wait == 0 {
		t.fail(key)
	}
	t.Unlock()
	if wait == 0 {
		return false
	}
	w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(wait.Seconds()))))
	respond.JSON(w, http.StatusTooManyRequests, respond.Failure{Error: t.message})
	return true
}
