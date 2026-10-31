---
name: notes-cli-command
description: Implement a notes-go CLI command (add/show/edit/remove on tag, note, folder) end to end - config.yaml entry, cobra wiring, a shared plum Process doing the bun query, its tests, and optional --yaml/--json output.
---

# notes-go CLI command

Use this skill to turn a stub in `notes-go/app/cli/cmd-<verb>-<noun>/cmd.go` into a working command.
Reference implementations: `createtag`, `deletetag`, `edittag`, `listtag` in `app/shared/services/process/`.
Process rules come from the `plum-process` skill; this skill only adds what is specific to notes-go.

## 1. Ask before writing

The user is learning bun: explain non-obvious choices, and ask (AskUserQuestion, recommended option first) about:

- **Input**: positional args (one or many) or flags. Many values -> `cobra.MinimumNArgs(1)`, a rename -> `cobra.ExactArgs(2)`.
- **Identifier**: value (case-insensitive, `lower(...)`) or uuid.
- **Edge cases**: already exists, not found, still referenced. Batch commands (create/delete) skip silently and return only affected rows; single-target commands (edit) return an error.

## 2. Layout

| What | Where |
|------|-------|
| Command name, help, flags | `config/config.yaml` under `commands.<verb>-<noun>` (struct in `config/config.go`) |
| Cobra wiring (`Args`, flags) | `cmd/<verb>.go` |
| CLI entry | `app/cli/cmd-<verb>-<noun>/cmd.go` |
| Business logic | `app/shared/services/process/<verb><noun>/run.go` (shared: the API routes reuse it) |
| Tests | same package, `run_test.go` |

Check the config entry exists: a missing key gives an empty `Use` and cobra silently cannot find the command.

## 3. Process

```go
type run struct {
	plumprimitive.Process
	cache

	ctx    context.Context
	db     bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	values []string
}

func (p *run) Run() {
	p.makeValues() // trim, drop blanks, SetErr(ErrNoValue) if none left
	p.insertTags() // guarded by if p.HasErr() { return }
}
```

- Take `bun.IDB`, never `*bun.DB`: callers can pass a transaction.
- Export sentinel errors (`ErrNoValue`, `ErrNotFound`, `ErrAlreadyExists`) and wrap with `%w` + the value, so callers use `errors.Is` and never import `pgdriver`.

## 4. bun patterns used here

- **Insert many, skip duplicates**: `Model(&tags).On("CONFLICT ((lower(value))) DO NOTHING").Returning("*").Scan(ctx, &result)`. The conflict target must match the unique index expression exactly; scan into a separate slice (bun would overwrite the first elements of `tags`).
- **Delete/update by value**: `Model((*domain.Tag)(nil))` (typed nil = table only). Lower both sides in SQL, not with `strings.ToLower`, so matching equals the index: `Where("lower(value) IN (SELECT lower(v) FROM unnest(?::text[]) AS v)", pgdialect.Array(values))`.
- **Not found**: `Scan` into a struct returns `sql.ErrNoRows` on 0 rows; into a slice it returns nil.
- **Unique violation**: let the index fail (no SELECT first, it races), then `errors.As(err, &pgErr) && pgErr.Field('C') == "23505"`.
- **List**: `Model(&result).OrderExpr("lower(value)").Scan(ctx)`.
- Deleting a parent relies on the `ON DELETE CASCADE` FKs from `createschema`: no code needed.

## 5. CLI entry

```go
func Create(cobra *cobra.Command, args []string) error {
	process := createtag.Process(cobra.Context(), postgrey.GetDb(), args)
	process.Run()

	if !process.HasErr() {
		fmt.Printf("Successfully created %d tag(s)\n", len(process.GetResult()))
	}

	return process.GetErr()
}
```

Check errors with `HasErr()`, never `if err := process.GetErr(); err != nil`. The DB comes from `postgrey.GetDb()` (set once in `cmd/root.go`).

## 6. Flags and --yaml / --json

Same pattern as `scne/develop/cmd/config.go`:

```yaml
# config/config.yaml
  read-tag:
    use: "tag"
    args:
      - { name: "list", shorthand: "l", value: "", usage: "List all tags", type: "bool" }
      - { name: "yaml", shorthand: "", value: "", usage: "Display the result as YAML", type: "bool" }
      - { name: "json", shorthand: "", value: "", usage: "Display the result as JSON", type: "bool" }
```

```go
// cmd/read.go init()
plumcliutilsargs.Makeflags(cmdReadTag, cmd.ReadTag.Flags)
cmdReadTag.MarkFlagsMutuallyExclusive("yaml", "json")

// cmd.go
var asList, asYAML, asJSON bool
if err := plumcliutilsargs.GetBoolFlags(cobra, map[string]*bool{"list": &asList, "yaml": &asYAML, "json": &asJSON}); err != nil {
	return err
}
// switch: asJSON -> plumdebug.Json(result), asYAML -> plumdebug.Yaml(result), default -> plain text
```

The printed domain struct needs `json:"-" yaml:"-"` on `bun.BaseModel` and lowercase `json`/`yaml` tags on fields.

## 6b. Flags + interactive input (reference: `add note`)

- Many-valued flag: `type: "stringSlice"` in config (plum `Makeflags`) -> `--tag a,b --tag c`, read with `flags.GetStringSlice`.
- Flags are resolved by a **mill** in `app/cli/cmd-<verb>-<noun>/services/mill/` (`Template{Ctx, Db, Flags}`, `Resolve()` returns the process `Input`), not by the process: the process stays UI-free.
- Missing values -> one `charm.land/huh/v2` form with only the missing fields (`NewInput`, `NewText`, `NewMultiSelect[string]().Filterable(true)` for `--x-from-list`); no field, no form.
- Mill test: build flags from `config.Current.Commands.<Cmd>.Flags` (catches renamed flags) and give every value, so no form/TTY is needed. Forms are checked by hand.
- Writes over several tables: `p.db.RunInTx` (a savepoint when `db` is a test `bun.Tx`), steps return errors inside it.
- `Returning("*")` fails on tables with DB-only columns (`items.parent_kind`): list the columns.

## 7. Tests

- **No-DB test**: blank input with `db = nil` -> `errors.Is(err, ErrNoValue)`. Also proves the `HasErr()` guard (a missing guard panics on nil).
- **Integration test**: connect via `postgrey.NewEnvironment()`, run `createschema`, `BeginTx` + `defer tx.Rollback()`, unique values `fmt.Sprintf("test-%d", time.Now().UnixNano())`, seed with other processes (`createtag`).
- A failed statement aborts the transaction: put the expected-DB-error case **last**.
- Run: `make test ARGS="./app/shared/services/process/..."` (needs Postgres; `colima start && docker compose up -d`).
- Then run the real command: `set -a && . ./.env && set +a && CONFIG_FILE=$PWD/config/config.yaml go run . <verb> <noun> ...`, and clean up what you created.
