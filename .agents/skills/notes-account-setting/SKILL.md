---
name: notes-account-setting
description: Add or change a setting stored on the account (users table) end to end - domain column (createschema adds it to existing bases), edituser validation, PATCH /api/me, OpenAPI and app types, the settings page, the layout sync, i18n and tests. Also use for any new column on an existing table.
---

# Setting of the account

The settings live as columns of `users`, read by `GET /api/me` and changed by `PATCH /api/me`: no settings table.
References: `locale`, `theme`, `color_mode`, `view`, `trash_retention_days`.

## 1. Go

| Step | Where |
|------|-------|
| Column | `notes-go/app/shared/domain/users.go`: field with `bun:",notnull,default:<v>"` and a snake_case `json` tag; the allowed values in an exported slice (`Locales`, `Themes`…), the first one the default |
| Validation + update | `app/shared/services/process/edituser/run.go`: a pointer field in `Input` (`json:"x,omitempty" binding:"optional"`), an `ErrInvalidX`, a case in `checkInput` (`valid()` for a choice), a `query.Set` in `updateUser` |
| Status | add `ErrInvalidX` to the 400 list of `app/api/services/respond/respond.go` (else the API answers 500) |
| Tests | `edituser/run_test.go` (invalid values without DB, defaults and saved values with `testaccount`), the `PATCH /api/me` checks of `app/api/router_test.go` |

Pitfall: bun inserts DEFAULT for a zero value when the tag has `default:`. A setting where 0 or "" is meaningful (retention 0 = never) is only set by an update: say so in a comment.
bun maps `int` to `bigint`: cast in SQL where Postgres wants an `int` (`make_interval(days => x::int)`).

## 2. Existing bases: added columns only

`createschema` adds a column of the model missing from an existing table (`ADD COLUMN IF NOT EXISTS`, with its type, NOT NULL and default) every time the API or the CLI starts, the server included: a new column with a default needs nothing by hand. Still by hand, after `make backup`, with `make pgsql QUERY="..."` (dev) on every base, the server's too:

- a NOT NULL column without a default on a table with rows (fill it first),
- a renamed or retyped column, a dropped NOT NULL, the `UPDATE`s that change the meaning of a column: one `BEGIN; … COMMIT;`, kept in the commit message.

A column left in the base but gone from the model is harmless: queries read the model's columns (`Returning("?Columns")`, never `*`).

## 3. API surface

`make test`, then `make openapi`, then `cd ../notes-app && bun run types`, then `make api` (the API container serves the old binary until then).

## 4. App

- `notes-app/app/layouts/default.vue` loads `me` (`useLazyAsyncData('me')`, shared with the settings page) and applies the account to the app: its `watch(me, …)` sets the locale, the theme colors (`appConfig.ui.colors`), `colorMode.preference` and `useView()`. A setting that some other control also changes (the sidebar mode button, the view toggle) is saved back by the second watch there.
- `notes-app/app/pages/settings.vue`: one `<section>` per group, a row with title, `text-xs text-muted` description and a `USelectMenu` (`value-key="value"`, `:search-input="false"`, `aria-label`) whose change sends `PATCH /api/me` and stores the answer in `me.value`.
- Texts in both `i18n/locales/fr.json` and `en.json`, under `settings`.
- `bun run typecheck && bun run test`, then an e2e in `test/e2e/settings.spec.ts` that sets the setting from the page, polls `GET /api/me`, and **sets the account back in a `finally`**: every run shares the e2e account.

## 5. Finish

The `notes-release` skill: README (Réglages section of `notes-app/README.md`), commits, tag.
