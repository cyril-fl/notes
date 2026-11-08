---
name: notes-issue-hunt
description: Find the GitHub issues Notes is missing - a full audit of notes-go and notes-app (bugs, security, data loss, timeouts, untested risky code), checked against the open issues and the `shortcut:` comments, then the new issues created in French. Use when the user says "cherche les issues manquantes", "analyse le code pour ajouter les issues", "audit", "qu'est-ce qui manque".
---

# Find the missing issues

"Analyse le code pour les issues" means a real audit of the code, not only a list of the TODOs. Run every step before reporting.

## 1. What is already known

```sh
gh issue list -R cyril-fl/notes --state all --limit 200
git grep -nE "shortcut:|TODO|FIXME" -- . ':!**/node_modules/**' ':!.claude/**' ':!.agents/**'
```

An open issue or a `shortcut:` comment that names its limit is already known: skip it, unless the real load (one account, 2-3 devices) already crosses that limit. A closed issue whose bug is back is a new finding.

## 2. Audit

Delegate the reading to the `notes-auditor` subagent (`.claude/agents/notes-auditor.md`): it reads a lot and returns only the findings. Give it the list from step 1. Without subagents, follow its checklist yourself.

## 3. Check each finding

- Each finding needs a concrete case: this input or situation gives this wrong result. Open the file and confirm the lines yourself, because the subagent can be wrong.
- Drop style taste and vague worries. A tiny fix (a stale README line) is not an issue: propose to fix it directly.

## 4. Report, then create

Report in French, grouped by **À corriger** / **Devrait être corrigé** / **Bonus**, numbered, each finding with its file:line, the problem, the fix and what happens if it is skipped. When the user asked for the issues, create them right away:

```sh
gh issue create -R cyril-fl/notes -l bug|enhancement -t "<titre en français>" -b "<le problème, avec les fichiers>

**Exemple** : <le cas concret>

**Piste** : <le plus petit fix>"
```

Label `bug` when something gives a wrong result today, `enhancement` for a missing test or a cleanup. Then give the issue numbers and say what was left out (not run, not read).
