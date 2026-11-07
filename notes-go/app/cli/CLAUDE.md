# CLI

- Commands, help and flags are declared in `config/config.yaml` (`commands.<verb>-<noun>`), wired in `cmd/<verb>.go`, run by `app/cli/cmd-<verb>-<noun>/cmd.go`. New command: the `notes-cli-command` skill.
- `cmd/resolve.go` `resolved(kind, run)` turns paths (`/Projets/Idées`) into uuids in the args and `--parent`, and `--content -` into stdin: commands only see uuids.
- Flags are read by a mill (`services/mill/`) that opens a huh form for what is missing; `noFlags()` drives the interactive path, and its test checks every flag of the config is declared.
- Edit commands change only what a flag asks for: `--pin` alone must not rewrite a note or a folder (see `cmd-edit-note`, `cmd-edit-folder`).
- `cmd/root.go` connects, creates the schema, purges the trash, then loads the account of the CLI session, except for the commands in `noAccount` (login, user add, reset-password).
- Output: plain text by default, `--json`/`--yaml` where declared; messages in English.
