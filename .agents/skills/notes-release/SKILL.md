---
name: notes-release
description: Close a piece of work or a version of Notes - one commit per change with the house message style, READMEs, then the checks, the annotated tag, the GitHub issues and the handoff delegated to the notes-release-checker subagent (Sonnet). Use when a feature or fix is done, or when the user says "tag", "release", "version", "close the issue".
---

# Release a version of Notes

Two halves: you make the commits and the READMEs (they need judgment), then the `notes-release-checker` subagent, on Sonnet, runs the checks, tags, handles the issues and writes the handoff. Never hand the commits to a subagent.

## 1. Commits (you)

- On `develop` (`git branch --show-current`: another agent may have switched it). Never push: the user runs `git push origin develop --follow-tags`.
- One commit per change, in English, lower case after the prefix: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `style:`. The message says what the user gets ("feat: pick a theme of colors from its thumbnail in the settings"), not the files. `Fixes #n` in the body for an issue.
- **Stage only your own lines.** Other agents share the checkout: `git status` first, and for each file you stage, read `git diff <file>`. A file that also holds changes you did not make (another agent's work, or several of your changes) is never `git add`-ed whole: stage only your hunks (`git diff <file>` cut to your hunks, then `git apply --cached`), or stop and ask the user.
- Before each commit, `git diff --cached` holds only this change, and nothing in it needs a file left out (a component, a type, a route of uncommitted work): otherwise the commit does not build on its own.
- Generated files (`notes-go/docs/swagger.yaml`, `notes-app/app/types/api.d.ts`) follow the same rule: regenerate them only when they hold nothing of another agent's, else stage your hunks.
- lefthook runs gofmt, Prettier and actionlint on staged files, and checks the prefix.
- READMEs in French, in their own `docs:` commit or with the change: `notes-go/README.md` (CLI, API), `notes-app/README.md` (the app).

## 2. Checks, tag, issues, handoff (notes-release-checker)

Once every commit is made, launch the `notes-release-checker` subagent with:

- the short shas of the release's commits;
- minor (a `feat:` in the release) or patch, and the tag message: what the user gets;
- the issues to close, each with what was done, and the issues to open for what is left (title, label, **À faire**);
- whether to run the e2e (`make up` must be running);
- what the handoff must say.

It never commits nor pushes. If it reports a failing check, fix it with a new commit (part 1), then launch it again.

Then report to the user in French: the commits, the tag, the issues closed and opened, and `git push origin develop --follow-tags`.
