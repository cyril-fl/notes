# API

- One folder per route, `app/api/<verb>-<noun>/route.go`, wired in `router.go`; account routes in `login/`.
- `router.go` wraps the mux in `guard` (Origin check against `--origin`) then `auth` (session cookie `notes_session`, the account into the ctx). Only `POST /api/login` works without a session.
- A handler decodes with `respond.Decode` (unknown fields refused, 10 MiB max), runs the process with `r.Context()`, then `respond.Error(w, process.GetErr())` or `respond.JSON`/`respond.List`.
- A new sentinel error of a process goes into the `statuses` table of `services/respond/respond.go`, or it answers 500.
- The swag comments above the handler (`@Summary`, `@Param body body <pkg>.Input true "body"`, `@Success`, `@Failure`, `@Router`) are the OpenAPI: fields are required unless tagged `binding:"optional"`. Then `make openapi` and `bun run types` in `notes-app/`.
- Tests: `router_test.go` drives the real router on the test base, logged in through `POST /api/login`.
- `notes serve` (`app/cli/cmd-serve/cmd.go`) cuts a request at 30 s read and 30 s write: a route that moves big bodies (import, export, image upload) must lift its own deadline with `http.NewResponseController(w)` (#26).
- In prod the API listens on `127.0.0.1:8080` behind an HTTPS proxy: `r.RemoteAddr` is the proxy for every client, never key anything per device on it (#28).
