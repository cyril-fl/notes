// Package users is the account routes an admin uses, and the self-service sign-up: the rest of an account's own
// routes are in login/.
package users

import (
	"net"
	"net/http"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/api/services/auth"
	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/api/services/throttle"
	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/password"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/addinvite"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/adduser"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createsession"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/deleteuser"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/silo"
	"github.com/cyril-fl/notes/notes-go/config"
	"github.com/uptrace/bun"
)

// pendingInvites keeps the invitations still usable: neither used nor expired.
const pendingInvites = "used_at IS NULL AND expires_at > current_timestamp"

// created counts the accounts created by address in the public mode: one address cannot fill the server.
var created = throttle.New("too many accounts created from this address, try again later")

// Account is an account as the admin panel lists it: the settings of someone else are none of its business.
type Account struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Admin     bool      `json:"admin"`
	CreatedAt time.Time `json:"created_at"`
}

// Version is the server's build, for the admin panel.
type Version struct {
	Version string `json:"version"`
}

// Mode tells the login and sign-up pages what the server allows.
type Mode struct {
	Mode string `json:"mode"` // one of domain.AddUserModes
	// Whether the sign-up page may create an account right now: the mode is public, the base has no account yet
	// (it is waiting for its host), or the invitation given is good.
	Allowed bool `json:"allowed"`
}

// Modes answers whether an account can be created, without a session: GET /api/add-user.
// @Summary  Whether an account can be created
// @Tags     account
// @Produce  json
// @Param    invite query string false "token of an invitation"
// @Success  200 {object} Mode
// @Failure  500 {object} respond.Failure
// @Router   /api/add-user [get]
func Modes(db bun.IDB, mode string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowed, err := allowed(r, db, mode, r.URL.Query().Get("invite"))
		if err != nil {
			respond.Error(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, Mode{Mode: mode, Allowed: allowed})
	}
}

// Signup creates an account without a session and logs it in: POST /api/signup. Allowed by an invitation, by the
// public mode, or on a base with no account yet, which the first account then runs. In the public mode an address
// waits longer after each account it creates.
// @Summary  Create an account and log it in
// @Tags     account
// @Accept   json
// @Produce  json
// @Param    body body adduser.Input true "body"
// @Success  200 {object} createsession.Result
// @Failure  400 {object} respond.Failure
// @Failure  403 {object} respond.Failure
// @Failure  409 {object} respond.Failure
// @Failure  429 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/signup [post]
func Signup(db bun.IDB, mode string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input adduser.Input
		if !respond.Decode(w, r, &input) {
			return
		}

		switch allowed, err := allowed(r, db, mode, input.Invite); {
		case err != nil:
			respond.Error(w, err)
			return
		case !allowed:
			forbidden(w, "this server does not take new accounts")
			return
		}
		address, _, _ := net.SplitHostPort(r.RemoteAddr)
		throttled := mode == domain.AddUserModes[1]
		if throttled && created.Blocked(w, address) {
			return
		}

		process := adduser.Process(r.Context(), db, input)
		process.Run()
		if process.HasErr() {
			if throttled {
				created.Forgive(address)
			}
			respond.Error(w, process.GetErr())
			return
		}

		// The account is created: it logs in as if it had just typed its name and password.
		session := createsession.Process(r.Context(), db, createsession.Input{Name: input.Name, Password: input.Password})
		session.Run()
		if session.HasErr() {
			respond.Error(w, session.GetErr())
			return
		}
		result := session.GetResult()
		auth.SetCookie(w, result.Token, result.ExpiresAt)
		respond.JSON(w, http.StatusOK, result)
	}
}

// List gives every account, for the admin panel: GET /api/users.
// @Summary  Every account
// @Tags     admin
// @Produce  json
// @Success  200 {array} Account
// @Failure  401 {object} respond.Failure
// @Failure  403 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/users [get]
func List(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAdmin(w, r, db) {
			return
		}

		var accounts []Account
		err := db.NewSelect().Table("users").Column("id", "name", "admin", "created_at").Order("name").Scan(r.Context(), &accounts)
		if err != nil {
			respond.Error(w, err)
			return
		}
		respond.List(w, http.StatusOK, accounts)
	}
}

// Delete deletes an account for good, with everything it holds: DELETE /api/users/{id}.
// @Summary  Delete an account and all its data
// @Tags     admin
// @Success  204
// @Failure  400 {object} respond.Failure
// @Failure  401 {object} respond.Failure
// @Failure  403 {object} respond.Failure
// @Failure  404 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/users/{id} [delete]
func Delete(db bun.IDB, store silo.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAdmin(w, r, db) {
			return
		}
		id, ok := respond.ID(w, r)
		if !ok {
			return
		}

		process := deleteuser.Process(r.Context(), db, store, id)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// Invite creates an invitation link: POST /api/invites. Its token is in the answer and nowhere else.
// @Summary  Create an invitation
// @Tags     admin
// @Produce  json
// @Success  201 {object} addinvite.Result
// @Failure  401 {object} respond.Failure
// @Failure  403 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/invites [post]
func Invite(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAdmin(w, r, db) {
			return
		}

		process := addinvite.Process(r.Context(), db)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.JSON(w, http.StatusCreated, process.GetResult())
	}
}

// Invites gives the invitations still usable: GET /api/invites.
// @Summary  The invitations still usable
// @Tags     admin
// @Produce  json
// @Success  200 {array} domain.Invite
// @Failure  401 {object} respond.Failure
// @Failure  403 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/invites [get]
func Invites(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAdmin(w, r, db) {
			return
		}

		// Each token is sealed with the key of the admin who made it, and an admin reads the invitations of the
		// others: their keys must be in memory before the scan opens the tokens.
		var admins []string
		if err := db.NewSelect().Model((*domain.Invite)(nil)).Distinct().Column("user_id").
			Where(pendingInvites).Scan(r.Context(), &admins); err != nil {
			respond.Error(w, err)
			return
		}
		for _, admin := range admins {
			if err := secret.Load(r.Context(), db, admin); err != nil {
				respond.Error(w, err)
				return
			}
		}

		var invites []domain.Invite
		err := db.NewSelect().Model(&invites).Where(pendingInvites).Order("created_at DESC").Scan(r.Context())
		if err != nil {
			respond.Error(w, err)
			return
		}
		respond.List(w, http.StatusOK, invites)
	}
}

// Revoke makes an invitation unusable: DELETE /api/invites/{id}.
// @Summary  Revoke an invitation
// @Tags     admin
// @Success  204
// @Failure  401 {object} respond.Failure
// @Failure  403 {object} respond.Failure
// @Failure  404 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/invites/{id} [delete]
func Revoke(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAdmin(w, r, db) {
			return
		}
		id, ok := respond.ID(w, r)
		if !ok {
			return
		}

		result, err := db.NewDelete().Model((*domain.Invite)(nil)).Where("id = ?", id).Exec(r.Context())
		if err != nil {
			respond.Error(w, err)
			return
		}
		if gone, err := result.RowsAffected(); err == nil && gone == 0 {
			respond.JSON(w, http.StatusNotFound, respond.Failure{Error: "no invitation with this id"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// Helpers

// allowed tells whether the sign-up page may create an account: see Mode.
func allowed(r *http.Request, db bun.IDB, mode, token string) (bool, error) {
	if mode == domain.AddUserModes[1] {
		return true, nil
	}

	taken, err := db.NewSelect().Model((*domain.User)(nil)).Exists(r.Context())
	if err != nil || !taken {
		return !taken, err
	}
	if token == "" {
		return false, nil
	}
	return db.NewSelect().Model((*domain.Invite)(nil)).
		Where("token_hash = ? AND used_at IS NULL AND expires_at > current_timestamp", password.TokenHash(token)).
		Exists(r.Context())
}

// Version is the server's build, for the admin to spot a mismatch with another deploy: GET /api/version.
// @Summary  The server's build version
// @Tags     admin
// @Produce  json
// @Success  200 {object} Version
// @Failure  401 {object} respond.Failure
// @Failure  403 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/version [get]
func VersionHandler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAdmin(w, r, db) {
			return
		}
		respond.JSON(w, http.StatusOK, Version{Version: config.Version()})
	}
}

// isAdmin answers 403 unless the account of the request may manage the others; on false the answer is written.
func isAdmin(w http.ResponseWriter, r *http.Request, db bun.IDB) bool {
	admin, err := db.NewSelect().Model((*domain.User)(nil)).Where("id = ? AND admin", domain.UserID(r.Context())).Exists(r.Context())
	if err != nil {
		respond.Error(w, err)
		return false
	}
	if !admin {
		forbidden(w, "this account does not manage the others")
		return false
	}
	return true
}

func forbidden(w http.ResponseWriter, message string) {
	respond.JSON(w, http.StatusForbidden, respond.Failure{Error: message})
}
