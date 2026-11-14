package login

import (
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/api/services/auth"
	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/countnotes"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createsession"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/deletesession"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/editpassword"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/edituser"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/purgeaccount"
	"github.com/uptrace/bun"
)

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
	return func(w http.ResponseWriter, r *http.Request) {
		address, _, _ := net.SplitHostPort(r.RemoteAddr)
		key := "address " + address
		var input createsession.Input
		if !respond.Decode(w, r, &input) {
			return
		}
		if failed.blocked(w, key) {
			return
		}

		process := createsession.Process(r.Context(), db, input)
		process.Run()
		if process.HasErr() {
			if !errors.Is(process.GetErr(), createsession.ErrBadCredentials) {
				failed.forgive(key)
			}
			respond.Error(w, process.GetErr())
			return
		}
		failed.reset(key)
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
// @Failure  429 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/me/password [patch]
func EditPassword(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := "account " + domain.UserID(r.Context())
		var input editpassword.Input
		if !respond.Decode(w, r, &input) {
			return
		}
		if failed.blocked(w, key) {
			return
		}

		process := editpassword.Process(r.Context(), db, input, auth.Token(r))
		process.Run()
		if process.HasErr() {
			if !errors.Is(process.GetErr(), editpassword.ErrWrongPassword) {
				failed.forgive(key)
			}
			respond.Error(w, process.GetErr())
			return
		}
		failed.reset(key)
		w.WriteHeader(http.StatusNoContent)
	}
}

// Stats counts the live notes of the account logged in and the bytes its notes take in the base: GET /api/me/stats.
// @Summary  Count the notes of the account logged in and their size in the base
// @Tags     account
// @Produce  json
// @Success  200 {object} countnotes.Stats
// @Failure  401 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/me/stats [get]
func Stats(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		process := countnotes.Process(r.Context(), db)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.JSON(w, http.StatusOK, process.GetResult())
	}
}

// DeleteData deletes for good every note, folder, tag and image of the account logged in: DELETE /api/me/data.
// The account, its settings and its sessions stay.
// @Summary  Delete every note, folder, tag and image of the account logged in
// @Tags     account
// @Success  204
// @Failure  401 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/me/data [delete]
func DeleteData(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		process := purgeaccount.Process(r.Context(), db)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
