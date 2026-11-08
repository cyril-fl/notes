---
name: notes-wiki
description: Bring the GitHub wiki of Notes (../notes.wiki) up to the version just tagged - a Versions.md section, the current version on Home.md, the pages the change touches (Application-web, Titres-et-tags, Thème, Données), screenshots of a new screen, then the wiki commit and the push left to the user. Use right after tagging a version (notes-release, notes-issue-done), or when the user says "wiki", "mets à jour le wiki", "doc utilisateur".
---

# Update the wiki

The wiki is a separate repo, `../notes.wiki` (from `develop/`), in French. It fell seven versions behind once (issue #44) because no step owned it: every tag now gets its wiki update, in the same session.

## Start from GitHub

```sh
cd ../notes.wiki
git status --short          # the user's own edits: keep them, never stage them
git fetch && git log --oneline -3 HEAD origin/master
```

Work on `origin/master`. If the local `master` is behind, `git pull --ff-only`; if it diverged, stop and ask. `.private/` is the user's, never commit it.

## What to write

Find what changed since the last version the wiki knows (`grep -m1 '^## v' Versions.md`): `git tag -l --format='%(contents)'` of each newer tag, and `git log --format='%h %s' <last>..HEAD` in `develop/` (skip `test:`, `chore:`, `style:` unless the user sees it).

- **`Versions.md`**: one section per tag, newest first, above the previous one: `## vX.Y.Z : <titre court> (<j> <mois abrégé> <année>)`, then bullets saying what the user gets. A manual step (an `ALTER`, an env var) goes in its bullet.
- **`Home.md`**: `Version actuelle : **vX.Y.Z**.`
- **The page of the feature**, so that it describes the app as it is now, not as a changelog:
  - `Application-web.md`: screens, shortcuts, settings (in their section), Pro mode.
  - `Titres-et-tags.md`: titles, tags, `[[links]]`.
  - `Thème.md`: themes, modes, fonts.
  - `Données.md`: a new column or table in the mermaid diagram and the table list, encryption.
  A new page also goes into `_Sidebar.md` and the page list of `Home.md`.
- Never: CLI commands, API routes or installation (they live in the READMEs), Tailscale addresses or the user's network setup.

Check each claim against the code or the i18n strings (`notes-app/i18n/locales/fr.json`) for the exact labels; the tag message alone is too short.

## Screenshots

A new screen or a visible change gets a PNG in `images/` (`app-<sujet>.png`), linked from its page with an alt text in French. Use the `notes-app-screenshot` skill. For a shot that needs content or a click:

- write a throwaway Playwright script in the session scratchpad, run from `notes-app/` with `NODE_PATH=$PWD/node_modules` (CommonJS `require`);
- create demo notes in a folder of the e2e account (`POST /api/folders`, `POST /api/notes` with `parent_id` and `content`), set the settings the shot needs (`PATCH /api/me`, e.g. `pro`), then in `finally` delete the folder (`DELETE /api/folders/{id}`, then `DELETE /api/trash/{id}`) and put back only the settings changed;
- hide the Nuxt devtools button before the shot: `page.addStyleTag({ content: '#nuxt-devtools-container{display:none!important}' })`;
- look at every PNG before copying it to `images/`.

## Commit

```sh
git add <pages> <images>    # explicit paths
git commit -m "docs: vX.Y.Z, <what the pages now cover>"
git log --oneline origin/master..master   # what the user still has to push
```

The user pushes, as for `develop`. An unpushed wiki commit is what left the online wiki behind: the report always lists the wiki commits not on `origin/master` and gives `git -C ../notes.wiki push origin master`.
