# Notes

A note-taking app, Obsidian-like: one Go program (`notes-go/`, CLI and HTTP API) and a Nuxt web app (`notes-app/`) on PostgreSQL, used through Tailscale from a phone and a tablet.

- `notes-go/`: CLI (`app/cli/`, `cmd/`), API (`app/api/`), the shared business logic (`app/shared/services/process/`, one plum Process each, used by both). See `notes-go/CLAUDE.md`.
- `notes-app/`: Nuxt 4 + Nuxt UI 4, client only (`ssr: false`). See `notes-app/CLAUDE.md`.
- `../deprecated/` is not the previous code: ignore it. `../notes.wiki/` is the GitHub wiki (separate repo).

## Run

- `make up` / `make down` at the root: Postgres and the API in Docker (Colima), the web app on :3000 (log in `.build/notes-app.log`).
- `make e2e`: the browser tests against that stack.
- CI (`.github/workflows/ci.yml`, on push and PR): gofmt, vet and the Go tests on a Postgres service; lint, Prettier, unit tests, typecheck and build of the app. Not the e2e.

## Working with the user

- Talk to the user in French. Code, comments and commit messages in English; READMEs and wiki in French.
- One commit per change (`feat:`/`fix:`/`refactor:`/`test:`/`docs:`/`chore:`/`style:`, checked by the lefthook `commit-msg` hook: `make setup` in `notes-go/`), an annotated tag per version, the user pushes. Details and issue handling: the `notes-release` skill.
- Work left for later goes into GitHub issues (`gh issue list -R cyril-fl/notes`), and known limits into `shortcut:` comments in the code.
- Before deleting or migrating data: `make backup` in `notes-go/`.
- Several agents share this checkout: files left uncommitted by another one are not yours (stage paths explicitly), and the branch may have been switched under you (`git branch --show-current` before committing; work goes on `develop`, `main` follows it at releases).
- `AGENTS.md` files are links to these `CLAUDE.md`: edit the `CLAUDE.md`.

## Skills

`notes-cli-command`, `notes-account-setting`, `notes-app-screenshot`, `notes-release`, `plum-process`, `go-tests` / `go-test-create` / `go-test-run`, `primitive-first-review`.
