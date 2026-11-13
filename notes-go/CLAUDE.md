# notes-go

Go module `github.com/cyril-fl/notes/notes-go` (`go install …/notes-go@latest`): the CLI (cobra, `cmd/` + `app/cli/`), the API (`notes serve`, `app/api/`) and the MCP server for agents (`notes mcp`, `app/cli/cmd-mcp/`) share the processes of `app/shared/services/process/` and the models of `app/shared/domain/`.

## Commands

- `make test` (`ARGS=./app/...` to narrow): runs in the base `<POSTGRES_DB>_test`, never the dev one. Needs Postgres up.
- `make api`: rebuilds and restarts the API container. The app talks to the old binary until then.
- `make openapi` after an API change (swag comments), then `bun run types` in `notes-app/`.
- `make backup` before touching data or the schema.
- The CLI by hand: `set -a && . ./.env && set +a && CONFIG_FILE=$PWD/config/config.yaml go run . <args>`. It acts for the account of the local CLI session: do not run `notes logout`, it deletes that session.

## Rules

- Processes follow the `plum-process` skill: guarded steps, `HasErr()` then `GetErr()`, `bun.IDB` so tests pass a transaction.
- Every query filters by account: `domain.UserID(ctx)`. A `bun.Tx` hands its `RunInTx` functions the ctx it began with, so begin transactions with a ctx that has the account; tests use `ctx, tx = testaccount.New(t, ctx, tx)`.
- **No migrations.** `createschema` creates missing tables, indexes and the columns a model gained (at every start, the server too). A rename, a type change or a NOT NULL column without a default needs a hand-written `ALTER` on every base (`make pgsql QUERY=...`, see the `notes-account-setting` skill). `RETURNING` lists the model's columns (`?Columns`), never `*`: an old column left in a base must not break a scan.
- Sentinel errors are exported and wrapped with `%w`; the API maps them to statuses in `app/api/services/respond/respond.go`.
- Periodic cleanups go into the `every` loop of `notes serve` (`app/cli/cmd-serve/cmd.go`, with `purgeitems` and `purgefiles`); the CLI only purges at startup (`cmd/root.go`).
- Images: a row in `files`, the bytes in Silo (S3, `compose.yaml`) through `services/silo.Store`, cited by notes as `/api/files/<id>`. Processes take the store as a parameter, tests pass `silo.NewMemory()`. `purgefiles` deletes what no note cites.
- Domain structs: `bun.BaseModel` gets `json:"-"`, fields snake_case `json` tags (they are the API's JSON and the app's types).
