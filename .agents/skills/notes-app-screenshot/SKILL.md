---
name: notes-app-screenshot
description: Look at a page of the running Notes web app as the e2e account, optionally with settings of the account changed for the shot (theme, color mode, language), restored afterwards. Use to check a UI change, a theme's contrast, light and dark mode, or before/after visuals.
---

# Screenshot the app

```sh
cd notes-app
node scripts/screenshot.mjs <out.png> [page path] [settings JSON for PATCH /api/me]
node scripts/screenshot.mjs "$SCRATCH/nord-dark.png" /settings '{"theme":"nord","color_mode":"dark"}'
```

Then read the PNG with the Read tool. Write the PNGs to the session scratchpad, never into the repo.

- `E2E_WIDTH=390` shoots at phone width (1100 by default).
- Needs the stack (`make up` at the root: API on :8080, app on :3000) and `test-results/.session.json`, written by the login of `bun run e2e` (run it once if the script says 401).
- The browser's color scheme does not matter: the account's `color_mode` wins. Pass `color_mode` to see light and dark.
- The script puts back only the keys it changed. Do not run it while `bun run e2e` runs: both use the e2e account.
- Several themes or modes: loop in the shell, one PNG each, then compare.
- For an interaction (a click, a menu open), write a throwaway Playwright test or extend the script locally; keep `storageState` and the restore in `finally`.
