---
name: notes-issue-done
description: Finish a GitHub issue of Notes without being asked - commit, tag a version, close the issue, open the follow-up issue for what is left. Use as soon as the work on an issue is done and its tests pass, or when the user says "issue finie", "ferme l'issue", "termine l'issue".
---

# Finish an issue

Run these steps in order, once the work on the issue is done and its tests pass. The details (message style, checks, commands) are in the `notes-release` skill: follow it.

1. **Commit**: on `develop`, paths staged explicitly, one commit per change, `Fixes #n` in the body.
2. **Version**: the next tag after `git tag --sort=-v:refname | head -1`. Bump the minor version (`v0.28.0`) if the issue adds a `feat:`. Otherwise (`fix:`, `refactor:`, `docs:`) bump the patch (`v0.27.2`). Make it an annotated tag: `git tag -a vX.Y.Z -m "vX.Y.Z: <what the user gets>"`.
3. **Wiki**: run the `notes-wiki` skill for the new tag (Versions, Home, the page of the feature, screenshots of a new screen), committed in the wiki.
4. **Close** the issue by hand, because `Fixes #n` closes nothing on `develop`. Do it even though nothing is pushed yet:
   ```sh
   gh issue close <n> -R cyril-fl/notes -c "Fait dans <short sha>, <tag> (sur develop) : <what changed>."
   ```
5. **Follow-up issue**: if part of the work is left, open one issue per remaining task. Write it in French, with **À faire** and `Suite de #n.`. Something left on the user's side counts too, such as an `ALTER` on the dev base or the server. Put that step in the issue and in the closing comment.
   ```sh
   gh issue create -R cyril-fl/notes -t "<titre>" -l enhancement -b "Suite de #<n>.

   **À faire**
   - ..."
   ```
6. **Report** to the user in French: the commit, the tag, the wiki pages updated, the closed issue, the new issue numbers, and what they must do (`git push origin develop --follow-tags`, `git -C ../notes.wiki push origin master`, a manual step). Never push.
