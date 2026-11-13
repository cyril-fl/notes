# API

- One folder per route, `app/api/<verb>-<noun>/route.go`, wired in `router.go`; an account's own routes in `login/`, the admin's and the sign-up in `users/`.
- `router.go` wraps the mux in `guard` (Origin check against `--origin`) then `auth` (session cookie `notes_session`, the account into the ctx). Only the routes of `auth`'s `public` map work without a session: the login, the sign-up and the mode it reads. An admin-only route checks the flag itself (`users.isAdmin`), the middleware knows nothing of it.
- A handler decodes with `respond.Decode` (unknown fields refused, 10 MiB max), runs the process with `r.Context()`, then `respond.Error(w, process.GetErr())` or `respond.JSON`/`respond.List`.
- A new sentinel error of a process goes into the `statuses` table of `services/respond/respond.go`, or it answers 500.
- The swag comments above the handler (`@Summary`, `@Param body body <pkg>.Input true "body"`, `@Success`, `@Failure`, `@Router`) are the OpenAPI: fields are required unless tagged `binding:"optional"`. Then `make openapi` and `bun run types` in `notes-app/`.
- Tests: `router_test.go` drives the real router on the test base, logged in through `POST /api/login`.
- `notes serve` (`app/cli/cmd-serve/cmd.go`) cuts a request at 30 s read and 30 s write: a route that moves big bodies (import, export, image upload) must lift its own deadline with `http.NewResponseController(w)` (#26).
- In prod the API listens on `127.0.0.1:8080` behind an HTTPS proxy: `r.RemoteAddr` is the proxy for every client, never key anything per device on it (#28).
