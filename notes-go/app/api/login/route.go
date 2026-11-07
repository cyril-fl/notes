package login

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/uptrace/bun"
	"main.go/app/api/services/auth"
	"main.go/app/api/services/respond"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/process/createsession"
	"main.go/app/shared/services/process/deletesession"
	"main.go/app/shared/services/process/editpassword"
	"main.go/app/shared/services/process/edituser"
)

// After maxFailures wrong passwords from one address, it waits lockout: a password cannot be guessed by trying them all.
// shortcut: kept in memory, a restart of the API forgets it; enough behind Tailscale, store it if the API ever faces the internet.
const (
	maxFailures = 5
	lockout     = 15 * time.Minute
)

type failures struct {
	sync.Mutex
	byAddress map[string][]time.Time
}

// recent is the failures of address within lockout, older ones dropped.
func (f *failures) recent(address string) int {
	f.Lock()
	defer f.Unlock()
	kept := f.byAddress[address][:0]
	for _, at := range f.byAddress[address] {
		if time.Since(at) < lockout {
			kept = append(kept, at)
		}
	}
	f.byAddress[address] = kept
	return len(kept)
}

func (f *failures) add(address string) {
	f.Lock()
	defer f.Unlock()
	f.byAddress[address] = append(f.byAddress[address], time.Now())
}

// Handler logs in: POST /api/login. The session token goes in an HttpOnly cookie, never in the body.
// @Summary  Log in
// @Tags     account
// @Accept   json
// @Produce  json
// @Param    body body createsession.Input true "body"
// @Success  200 {object} createsession.Result
// @Failure  400 {object} respond.Failure
// @Failure  401 {object} respond.Failure
// @Failure  429 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/login [post]
func Handler(db bun.IDB) http.HandlerFunc {
	failed := &failures{byAddress: map[string][]time.Time{}}

	return func(w http.ResponseWriter, r *http.Request) {
		address, _, _ := net.SplitHostPort(r.RemoteAddr)
		if failed.recent(address) >= maxFailures {
			respond.JSON(w, http.StatusTooManyRequests, respond.Failure{Error: "too many failed logins, try again later"})
			return
		}

		var input createsession.Input
		if !respond.Decode(w, r, &input) {
			return
		}

		process := createsession.Process(r.Context(), db, input)
		process.Run()
		if process.HasErr() {
			failed.add(address)
			respond.Error(w, process.GetErr())
			return
		}
		result := process.GetResult()
		auth.SetCookie(w, result.Token, result.ExpiresAt)
		respond.JSON(w, http.StatusOK, result)
	}
}

// Logout logs this device out: POST /api/logout.
// @Summary  Log out
// @Tags     account
// @Success  204
// @Failure  401 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/logout [post]
func Logout(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		process := deletesession.Process(r.Context(), db, auth.Token(r))
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		auth.SetCookie(w, "", time.Time{})
		w.WriteHeader(http.StatusNoContent)
	}
}

// Me is the account logged in: GET /api/me.
// @Summary  The account logged in
// @Tags     account
// @Produce  json
// @Success  200 {object} domain.User
// @Failure  401 {object} respond.Failure
// @Router   /api/me [get]
func Me(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user domain.User
		if err := db.NewSelect().Model(&user).Where("id = ?", domain.UserID(r.Context())).Scan(r.Context()); err != nil {
			respond.Error(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, user)
	}
}

// EditMe changes the settings of the account logged in: PATCH /api/me.
// @Summary  Change the settings of the account logged in
// @Tags     account
// @Accept   json
// @Produce  json
// @Param    body body edituser.Input true "body"
// @Success  200 {object} domain.User
// @Failure  400 {object} respond.Failure
// @Failure  401 {object} respond.Failure
// @Failure  413 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/me [patch]
func EditMe(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input edituser.Input
		if !respond.Decode(w, r, &input) {
			return
		}

		process := edituser.Process(r.Context(), db, input)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.JSON(w, http.StatusOK, process.GetResult())
	}
}

// EditPassword changes the password of the account logged in, and logs its other devices out: PATCH /api/me/password.
// @Summary  Change the password of the account logged in
// @Tags     account
// @Accept   json
// @Param    body body editpassword.Input true "body"
// @Success  204
// @Failure  400 {object} respond.Failure
// @Failure  401 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/me/password [patch]
func EditPassword(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input editpassword.Input
		if !respond.Decode(w, r, &input) {
			return
		}

		process := editpassword.Process(r.Context(), db, input, auth.Token(r))
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
