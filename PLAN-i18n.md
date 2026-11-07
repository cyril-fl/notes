# Plan : i18n de l'app web (français / anglais)

But : l'interface de `notes-app` existe en français et en anglais. La langue est un réglage du compte, stocké dans
Postgres (table `users`), modifiable depuis la page Réglages.

Décisions déjà prises (ne pas les rediscuter) :
- Pas de deuxième base : une colonne `locale` sur `users`.
- Langues : `fr` (défaut) et `en`.
- Les messages d'erreur de l'API restent en anglais ; le front affiche ses propres textes (comme `conflict` dans `useApi.ts`).
- Le CLI n'est pas concerné.
- Ignorer `deprecated/`.

Étapes dans l'ordre. Une étape = un commit (n'indexer que ses propres fichiers, par nom).

---

## 1. Back : colonne `locale`

`notes-go/app/shared/domain/users.go`
- Ajouter à `User` : `Locale string \`bun:",notnull,default:'fr'" json:"locale"\``.
- Ajouter la liste des langues acceptées : `var Locales = []string{"fr", "en"}`.

Pas de migration (supprimées par `d58646b`) : `CREATE TABLE` crée la colonne sur une base neuve. Sur une base existante,
une fois à la main :
```sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS locale varchar NOT NULL DEFAULT 'fr'
```

Vérif : `make test` dans `notes-go` (base `<POSTGRES_DB>_test`).

## 2. Back : process `editlocale`

Nouveau dossier `notes-go/app/shared/services/process/editlocale/` (`run.go` + `run_test.go`), même forme que
`process/edittag` (struct `run` qui embarque `plumprimitive.Process` + `cache`, `Process(ctx, db, locale)`, `Run()`, `GetResult()`).
- `ErrInvalidLocale = errors.New("unknown locale")`.
- `checkLocale` : refuse ce qui n'est pas dans `domain.Locales` (`slices.Contains`) → `ErrInvalidLocale`.
- `updateUser` : `UPDATE users SET locale = ? WHERE id = domain.UserID(ctx) RETURNING *`, résultat = `domain.User`.
- Tests : une locale invalide échoue sans DB (db nil, comme `TestBlankValues` dans edittag) ; `en` est enregistré puis relu
  (transaction annulée + `testaccount`, comme `TestRename`).

`notes-go/app/api/services/respond/respond.go` : ajouter `editlocale.ErrInvalidLocale` au groupe `http.StatusBadRequest`.

## 3. Back : route `PATCH /api/me`

`notes-go/app/api/login/route.go`, à côté de `Me` :
- `type MeBody struct { Locale string \`json:"locale"\` }`.
- `EditMe(db bun.IDB) http.HandlerFunc` : `respond.Decode` → `editlocale.Process(...).Run()` → `respond.Error` ou
  `respond.JSON(w, 200, process.GetResult())`. Commentaires swag comme `edittag.Handler` (`@Tags account`,
  `@Router /api/me [patch]`, 200 `domain.User`, 400/401/500).

`notes-go/app/api/router.go` : `mux.Handle("PATCH /api/me", apilogin.EditMe(db))`, sous `GET /api/me`.

Puis :
- `cd notes-go && make openapi` (régénère `docs/swagger.yaml`)
- `cd notes-app && pnpm types` (régénère `app/types/api.d.ts` ; `User` aura `locale`)

Vérif : `make test` (dont `app/api/router_test.go`).

## 4. Front : installer i18n

- `pnpm add @nuxtjs/i18n` (version compatible Nuxt 4, v10+).
- `nuxt.config.ts` : ajouter le module et
  ```ts
  i18n: {
    strategy: 'no_prefix', // pas de /en/ dans les URL
    defaultLocale: 'fr',
    locales: [{ code: 'fr', file: 'fr.json' }, { code: 'en', file: 'en.json' }],
    // Le cookie garde la dernière langue choisie : la page de login, sans compte, l'affiche aussi.
    detectBrowserLanguage: { useCookie: true, cookieKey: 'notes-locale' },
  },
  ```
- Créer `notes-app/i18n/locales/fr.json` et `en.json` (dossier par défaut du module en v10).
- `app/app.vue` : remplacer `:locale="fr"` par la locale Nuxt UI courante :
  ```ts
  import { en, fr } from '@nuxt/ui/locale';
  const { locale } = useI18n();
  const uiLocale = computed(() => (locale.value === 'en' ? en : fr));
  ```

## 5. Front : extraire les textes

Remplacer chaque texte en dur par `t('…')` (`const { t } = useI18n()`, ou `$t` dans le template). Clés groupées par
écran : `sidebar.*`, `settings.*`, `trash.*`, `items.*`, `api.*`… Remplir `fr.json` avec le texte actuel, `en.json`
avec sa traduction.

Fichiers concernés :
- composants : `AppSearch`, `AppSidebar`, `ConfirmDialog`, `FolderNode`, `FolderView`, `GraphView`, `NoteEditor`,
  `PathNav`, `PromptDialog`, `RelatedNotes`, `ViewToggle`, `ItemRow`
- pages : `index`, `login`, `search`, `settings`, `trash`, `graph`, `folders/[id]`, `tags/[value]`
- composables : `useApi.ts` (messages d'erreur + `conflict` passés par les appelants), `useItems.ts` (menus, prompts, « Racine »)
- `utils/title.ts` : `'Sans titre'` est testé par `node --test` sans Nuxt : ne pas y appeler `useI18n`.
  Ajouter un paramètre `fallback` à `displayTitle` que les appelants remplissent avec `t('untitled')`, et adapter les tests.

Points particuliers :
- Pluriels : `'1 note(s) et 1 dossier(s) importés'` → message pluriel vue-i18n (`"{n} note | {n} notes"`).
- Dates : `toLocaleDateString('fr-FR', …)` dans `ItemRow.vue` et `trash.vue` → utiliser `locale.value`.
- Contrôle final : `grep -rnE "[éèàçêù]" app` ne doit plus rien trouver en dehors des commentaires.

## 6. Front : appliquer la langue du compte + réglage

- `app/layouts/default.vue` (toutes les pages connectées ; login a `layout: false`) : lire `/api/me` avec la même clé que
  la page Réglages (`useLazyAsyncData('me', …)`), et `setLocale(me.locale)` quand il arrive.
- `app/pages/settings.vue` : nouvelle section « Langue » entre Compte et Données : un `USelectMenu`
  (`Français` / `English`, `value-key="value"`, `aria-label`). Au changement : `api<User>('/api/me', { method: 'PATCH', body: { locale } })`,
  puis `setLocale(locale)` et mise à jour de `me`.

## 7. Tests

- `pnpm test` et `pnpm typecheck` passent.
- e2e (`test/e2e/settings.spec.ts`) : passer en English dans Réglages → un libellé est en anglais (ex. le titre de page),
  recharger → toujours en anglais (la langue vient de l'API), remettre en français à la fin. Les specs e2e existantes
  cherchent des textes français : elles doivent continuer à passer avec `fr` par défaut.
- `pnpm e2e` passe.

## Hors périmètre

Langue du CLI, traduction des messages d'erreur de l'API, troisième langue, détection de langue côté Go.
