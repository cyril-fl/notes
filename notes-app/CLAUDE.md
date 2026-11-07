# notes-app

Nuxt 4, Nuxt UI 4 (Tailwind 4), `@nuxtjs/i18n`, VueUse; client only, the Go API on `runtimeConfig.public.apiBase`.

## Commands

`pnpm test` (node unit tests in `test/*.test.ts`), `pnpm typecheck`, `pnpm e2e` (Playwright in the installed Chrome, against `make up`), `pnpm types` (API types from `../notes-go/docs/swagger.yaml`).

## Rules

- API calls go through `useApi()`: cookies included, toast on errors, `/login` on a 401. Types come from `app/types/items.ts` (aliases of the generated `api.d.ts`), never written by hand.
- Every text is in `i18n/locales/fr.json` **and** `en.json`; French is the default.
- Icons are named once in `composables/useIcons.ts`.
- Use Nuxt UI components and its tokens (`text-muted`, `bg-elevated`, `ring-primary`…), not raw colors: the themes change them.
- The account drives the app: `layouts/default.vue` loads `me` and applies locale, theme, mode and view. Settings: the `notes-account-setting` skill.
- Themes: palettes in `app/assets/css/main.css` (`@theme static`, an accent and a `-gray` scale, the greys serving light at 50 and dark at 900), the list in `app/utils/themes.ts`, the ids repeated in `domain.Themes` (Go), the names in i18n `settings.themes`. Check both modes with the `notes-app-screenshot` skill.

## e2e

- `test/e2e/login.ts` logs the e2e account in once (`E2E_NAME`/`E2E_PASSWORD`, default `e2e`/`e2e-password`) into `test-results/.session.json`; the browser locale is `fr-FR`, so selectors read French.
- `sandbox()` of `helpers.ts` gives a test its own folder, deleted afterwards.
- A test that changes the account (settings) sets it back in a `finally`: every run shares that account.
