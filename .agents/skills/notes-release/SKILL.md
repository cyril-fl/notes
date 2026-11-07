---
name: notes-release
description: Close a piece of work or a version of Notes - one commit per change with the house message style, READMEs, annotated tag, GitHub issues closed by hand (develop is not the default branch), the handoff for the next session. Use when a feature or fix is done, or when the user says "tag", "release", "version", "close the issue".
---

# Release a version of Notes

## Commits

- One commit per change, in English, lower case after the prefix: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `style:`. The message says what the user gets ("feat: pick a theme of colors from its thumbnail in the settings"), not the files.
- `Fixes #n` in the body for an issue.
- On `develop` (`git branch --show-current`: another agent may have switched it). Stage paths explicitly; never `git add` files another agent left uncommitted (check `git status` first).
- lefthook runs gofmt on staged Go files.
- Never push: the user runs `git push origin develop --follow-tags`.

## Before the tag

- `cd notes-go && make test`; `cd notes-app && bun run test && bun run typecheck && bun run e2e` (needs `make up`).
- READMEs in French: `notes-go/README.md` (CLI, API), `notes-app/README.md` (the app).
- The wiki is a separate repo, `../notes.wiki` (pages `Home.md`, `cli.md`, `api.md`): check `git status` there, it may hold the user's own uncommitted edits.

## Tag

- `git tag -a vX.Y.0 -m "vX.Y.0: <what the user gets>"`. Check `git tag --sort=-v:refname | head` first: another agent may already have used the next number.
- A tag not yet pushed (`git branch -r --contains` / not in `origin/develop`) may be moved with `git tag -fa`; a pushed one never.

## Issues

`gh issue list -R cyril-fl/notes`. `develop` is not the default branch: `Fixes #n` does not close anything. Close by hand once the work is committed:

```sh
gh issue close <n> -R cyril-fl/notes -c "Fait dans <short sha> (sur develop) : <what changed>."
```

Obsolete issues: `-r "not planned"` with the reason. Work left for later goes into a new issue (labels `bug`, `enhancement`, `documentation`, `question`), in French, with **À faire** or **Piste**.

## Handoff

The user switches sessions to save tokens: update the « Reprise » section of the plan file named in memory, and the roadmap memory, with what was done, the tag, and what is left (as issue numbers).
