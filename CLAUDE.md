# Notes

A note-taking app, Obsidian-like: one Go program (`notes-go/`, CLI and HTTP API) and a Nuxt web app (`notes-app/`) on PostgreSQL, used through Tailscale from a phone and a tablet.

- `notes-go/`: CLI (`app/cli/`, `cmd/`), API (`app/api/`), the shared business logic (`app/shared/services/process/`, one plum Process each, used by both). See `notes-go/CLAUDE.md`.
- `notes-app/`: Nuxt 4 + Nuxt UI 4, client only (`ssr: false`). See `notes-app/CLAUDE.md`.
- `../deprecated/` is not the previous code: ignore it. `../notes.wiki/` is the GitHub wiki (separate repo).

## Run

- `make up` / `make down` at the root: Postgres and the API in Docker (Colima), the web app on :3000 (log in `.build/notes-app.log`).
- `make e2e`: the browser tests against that stack.

## Working with the user

- Talk to the user in French. Code, comments and commit messages in English; READMEs and wiki in French.
- One commit per change (`feat:`/`fix:`/`refactor:`/`test:`/`docs:`), an annotated tag per version, the user pushes. Details and issue handling: the `notes-release` skill.
- Work left for later goes into GitHub issues (`gh issue list -R cyril-fl/notes`), and known limits into `shortcut:` comments in the code.
- Before deleting or migrating data: `make backup` in `notes-go/`.
- Files left uncommitted by another agent are not yours: stage paths explicitly.

## Skills

`notes-cli-command`, `notes-account-setting`, `notes-app-screenshot`, `notes-release`, `plum-process`, `go-tests` / `go-test-create` / `go-test-run`, `primitive-first-review`.
