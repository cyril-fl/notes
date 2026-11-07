# API

- One folder per route, `app/api/<verb>-<noun>/route.go`, wired in `router.go`; account routes in `login/`.
- `router.go` wraps the mux in `guard` (Origin check against `--origin`) then `auth` (session cookie `notes_session`, the account into the ctx). Only `POST /api/login` works without a session.
- A handler decodes with `respond.Decode` (unknown fields refused, 10 MiB max), runs the process with `r.Context()`, then `respond.Error(w, process.GetErr())` or `respond.JSON`/`respond.List`.
- A new sentinel error of a process goes into the `statuses` table of `services/respond/respond.go`, or it answers 500.
- The swag comments above the handler (`@Summary`, `@Param body body <pkg>.Input true "body"`, `@Success`, `@Failure`, `@Router`) are the OpenAPI: fields are required unless tagged `binding:"optional"`. Then `make openapi` and `pnpm types` in `notes-app/`.
- Tests: `router_test.go` drives the real router on the test base, logged in through `POST /api/login`.
