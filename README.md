# Notes — Mini Notion-like self-hosted

Application de prise de notes orientée simplicité, productivité et self-host.
Construite avec Nuxt 4, Vue 3, TipTap, MongoDB et Docker.

---

## Analyse produit

### 1. Fonctionnalités déjà présentes

| Domaine          | Fonctionnalité                                                       | État        |
| ---------------- | -------------------------------------------------------------------- | ----------- |
| **Éditeur**      | Éditeur rich text (TipTap) avec Markdown                             | Fonctionnel |
| **Éditeur**      | Toolbar : bold, italic, underline, strike, highlight, headings H1-H6 | Fonctionnel |
| **Éditeur**      | Listes (bullets, ordonnées, tâches/checklists)                       | Fonctionnel |
| **Éditeur**      | Insertion de liens avec validation URL                               | Fonctionnel |
| **Éditeur**      | Blocs de code, séparateurs horizontaux                               | Fonctionnel |
| **Éditeur**      | Undo/Redo                                                            | Fonctionnel |
| **Recherche**    | Recherche globale full-text (MiniSearch, fuzzy + prefix)             | Fonctionnel |
| **Recherche**    | Barre de recherche avec debounce et filtre par type (note/dossier)   | Fonctionnel |
| **Recherche**    | Page de résultats de recherche dédiée (`/search`)                    | Fonctionnel |
| **Recherche**    | Index de recherche synchronisé en temps réel (CRUD)                  | Fonctionnel |
| **Organisation** | Dossiers hiérarchiques (arbre récursif)                              | Fonctionnel |
| **Organisation** | Création de dossier dans un dossier                                  | Fonctionnel |
| **Organisation** | Création de note dans un dossier                                     | Fonctionnel |
| **Organisation** | Renommage de dossier (inline)                                        | Fonctionnel |
| **Organisation** | Suppression (note, dossier + descendants)                            | Fonctionnel |
| **Organisation** | Corbeille : soft delete, restauration, purge, vidage                 | Fonctionnel |
| **Organisation** | Création de dossier depuis la sidebar                                | Fonctionnel |
| **Navigation**   | Sidebar rétractable et redimensionnable                              | Fonctionnel |
| **Navigation**   | Arbre de navigation dossiers (tree view)                             | Fonctionnel |
| **Navigation**   | Menu contextuel (clic droit)                                         | Fonctionnel |
| **Navigation**   | Extraction automatique de hashtags                                   | Fonctionnel |
| **Navigation**   | Top 5 hashtags populaires dans la sidebar                            | Fonctionnel |
| **UI/UX**        | Dark/Light mode (system preference)                                  | Fonctionnel |
| **UI/UX**        | Design system complet (Nuxt UI + Tailwind)                           | Fonctionnel |
| **UI/UX**        | i18n (français)                                                      | Fonctionnel |
| **Backend**      | API REST CRUD complète                                               | Fonctionnel |
| **Backend**      | Validation Zod sur toutes les entrées                                | Fonctionnel |
| **Backend**      | Updates optimistes avec rollback                                     | Fonctionnel |
| **Backend**      | Connexion MongoDB directe (singleton)                                | Fonctionnel |
| **Backend**      | API de recherche full-text (`/api/data/search`)                      | Fonctionnel |
| **Éditeur**      | Auto-save temps réel avec debounce (1500ms)                          | Fonctionnel |
| **UI/UX**        | Notifications d'erreur (toasts) sur toutes les opérations CRUD       | Fonctionnel |
| **UI/UX**        | Modal de confirmation avant suppression                              | Fonctionnel |
| **UI/UX**        | Page d'erreur custom (404, erreurs serveur)                          | Fonctionnel |
| **Navigation**   | Page listing toutes les notes (`/notes`)                             | Fonctionnel |
| **UI/UX**        | Toolbar layout responsive (composant header réutilisable)            | Fonctionnel |
| **UI/UX**        | Système de configuration de page dynamique (`usePageSection`)        | Fonctionnel |
| **Infra**        | Docker multi-stage + Docker Compose                                  | Fonctionnel |
| **Infra**        | MongoDB persistant                                                   | Fonctionnel |
| **DX**           | ESLint, Prettier, Husky, CommitLint                                  | Fonctionnel |

---

### 2. Fonctionnalités essentielles manquantes

#### CORE INDISPENSABLE (pré-requis pour un usage quotidien)

| #   | Fonctionnalité                   | Justification                                                                                           | Complexité | État |
| --- | -------------------------------- | ------------------------------------------------------------------------------------------------------- | ---------- | ---- |
| C4  | **Drag & drop pour réorganiser** | Déplacer une note/dossier nécessite actuellement de supprimer et recréer. Essentiel pour l'organisation | Haute      | Non  |

#### AMÉLIORATION UX (qualité perçue, rétention)

| #   | Fonctionnalité                                   | Justification                                                                                            | Complexité | État        |
| --- | ------------------------------------------------ | -------------------------------------------------------------------------------------------------------- | ---------- | ----------- | --- | --- | ------ | --- |
| U1  | **Recherche par hashtag** (clic → filtre)        | Hashtags affichés en texte brut, non cliquables. L'API search supporte le filtrage par tags côté serveur | Faible     | Placeholder |
| U2  | **Tri et filtres** (date, titre, récents)        | Aucun moyen de trier les notes dans un dossier                                                           | Faible     | Non         |
| U3  | **Note épinglée / favoris**                      | Pattern standard pour accès rapide aux notes fréquentes                                                  | Moyenne    | Non         |
| U4  | **Breadcrumb de navigation**                     | Le `path` existe dans le modèle mais n'est pas affiché. L'utilisateur se perd dans l'arbre               | Faible     | Non         |
| U5  | **Raccourcis clavier globaux**                   | Cmd+N (nouvelle note), Cmd+K (recherche), Cmd+S (save explicite)                                         | Moyenne    | Non         |     |     | Faible | Non |
| U7  | **Responsive / Mobile**                          | Toolbar responsive basique (`sm:grid-cols`), mais pas de layout mobile dédié (hamburger, drawer)         | Moyenne    | Minimal     |
| U8  | **Persistance de l'état sidebar** (localStorage) | La largeur du panel est en mémoire uniquement, perdue au rechargement                                    | Faible     | Non         |

#### FONCTIONNALITÉS AVANCÉES (différenciation, v1+)

| #   | Fonctionnalité                       | Justification                                                                                      | Complexité | État        |
| --- | ------------------------------------ | -------------------------------------------------------------------------------------------------- | ---------- | ----------- |
| A1  | **Authentification basique**         | Aucune protection — critique pour le self-host exposé sur Internet                                 | Haute      | Non         |
| A2  | **Graph view** (liens entre notes)   | Route `/graph` = placeholder (`test`), lien commenté dans la nav, libs installées mais inutilisées | Haute      | Placeholder |
| A3  | **Images / upload de fichiers**      | Pas de support média dans l'éditeur                                                                | Haute      | Non         |
| A4  | **Export** (Markdown, PDF, HTML)     | Aucun moyen d'extraire ses données                                                                 | Moyenne    | Non         |
| A5  | **Import** (Markdown, .md files)     | Aucun moyen d'importer du contenu existant                                                         | Moyenne    | Non         |
| A6  | **Corbeille / soft delete**          | Soft delete avec `deletedAt`, corbeille UI, restauration, purge auto 30j                           | Moyenne    | Fonctionnel |
| A7  | **Historique / versioning de notes** | Pas de possibilité de revenir en arrière sur le contenu                                            | Haute      | Non         |
| A8  | **Multi-langue** (en + fr)           | Seul le français est supporté, limite l'adoption                                                   | Faible     | Non         |
| A9  | **PWA / offline**                    | Pas de service worker, inutilisable hors connexion                                                 | Haute      | Non         |
| A10 | **API de santé** (`/api/health`)     | Nécessaire pour monitoring Docker/orchestration                                                    | Faible     | Non         |

---

### 3. Hypothèses formulées

- **Auto-save** : sauvegarde automatique avec debounce 1500ms pendant la saisie, + sauvegarde sur `onBeforeRouteLeave` et `beforeunload`. Notifications d'erreur en cas d'échec.
- **Graph view** : les dépendances (graphology, sigma) sont installées et la route existe, mais le template contient uniquement `test`. Le lien dans la nav est commenté → placeholder.
- **Suppression** : delete cascade sur les dossiers (`getRelatedIds`) avec modal de confirmation. Corbeille avec soft delete, restauration et purge.
- **Hashtags** : l'extraction et l'affichage fonctionnent, et l'API search supporte le filtrage par tags, mais les hashtags sont rendus en texte brut (non cliquables).
- **Pas de tests** : aucun fichier de test détecté (ni unitaire, ni e2e). Manque significatif mais non bloquant pour le versioning initial.
- **MongoDB direct** : connexion MongoDB directe (singleton) en plus de Nitro Storage, utilisée par le moteur de recherche.

---

### 4. Estimation de version actuelle

| Critère                 | Évaluation                            |
| ----------------------- | ------------------------------------- |
| Structure de l'app      | Solide, bien architecturée            |
| CRUD complet            | Oui                                   |
| Éditeur fonctionnel     | Oui                                   |
| Navigation hiérarchique | Oui                                   |
| Recherche               | Oui (full-text, fuzzy, filtres)       |
| Auth                    | Non                                   |
| Tests                   | Non                                   |
| Stabilité API           | En cours de définition                |
| Production-ready        | Non (pas d'auth, pas de health check) |

#### Version recommandée : `0.5.0`

**Justification :**

- `0.x.x` → Pas encore stable / pas de contrat d'API public
- `0.3.0` correspondait aux fondations : structure initiale, éditeur + CRUD, navigation + dossiers + Docker
- `0.4.0` correspondait au milestone **Recherche** : recherche globale full-text, barre de recherche, page résultats, page notes
- `0.5.0` reflète le milestone **Fiabilité** atteint : auto-save debounce, confirmation de suppression, notifications d'erreur (toasts), page d'erreur custom, corbeille avec restauration/purge
- `1.0.0` sera atteint quand : auth + tests minimaux + drag & drop seront en place

---

### 5. Stratégie de versioning (SemVer)

```
MAJOR.MINOR.PATCH  →  0.5.0
```

#### Roadmap suggérée

| Version  | Milestone              | Contenu clé                                                                            |
| -------- | ---------------------- | -------------------------------------------------------------------------------------- |
| `0.3.0`  | **Fondations**         | CRUD, éditeur TipTap, dossiers, Docker                                                 |
| `0.4.0`  | **Recherche**          | Recherche globale full-text, barre de recherche, page résultats, page notes, toolbar   |
| `0.5.0`  | **Actuel — Fiabilité** | Auto-save debounce, confirmation delete, toasts d'erreur, page 404, corbeille          |
| `0.6.0`  | **Organisation**       | Drag & drop, tri/filtres, favoris, hashtags cliquables, raccourcis clavier, breadcrumb |
| `0.7.0`  | **Self-host ready**    | Auth basique (password ou token), health check API, responsive mobile                  |
| `0.8.0`  | **Contenu riche**      | Upload images, export MD/PDF, import MD                                                |
| `0.9.0`  | **Avancé**             | Graph view, multi-langue                                                               |
| `0.10.0` | **Qualité**            | Tests unitaires + e2e, PWA/offline, historique de notes                                |
| `1.0.0`  | **Stable**             | API stabilisée, documentation, première release publique                               |

#### Règles de versioning

- **PATCH** (`0.3.1`) : bugfix, correction CSS, typo i18n
- **MINOR** (`0.4.0`) : nouvelle fonctionnalité, nouvelle route API
- **MAJOR** (`1.0.0`) : première release stable, contrat d'API public
- Tant qu'on est en `0.x`, les breaking changes sont autorisés en MINOR

---

### 6. Où et comment ajouter le numéro de version

#### a) `package.json` (source de vérité)

```json
{
  "name": "notes",
  "version": "0.5.0"
}
```

#### b) UI — Footer ou page Settings/About

Récupérable via `runtimeConfig` :

```ts
// nuxt.config.ts
runtimeConfig: {
  public: {
    appVersion: process.env.npm_package_version || '0.0.0';
  }
}
```

Puis dans un composant :

```vue
<template>
  <span class="text-xs text-muted">v{{ version }}</span>
</template>
<script setup>
const { appVersion: version } = useRuntimeConfig().public;
</script>
```

#### c) API — Header ou endpoint

Option 1 : Header `X-App-Version` sur chaque réponse (via Nitro middleware)

Option 2 : Endpoint `/api/health` retournant :

```json
{ "status": "ok", "version": "0.5.0", "uptime": 12345 }
```

#### d) Docker — Labels & metadata

```dockerfile
LABEL org.opencontainers.image.version="0.3.0"
LABEL org.opencontainers.image.title="Notes"
LABEL org.opencontainers.image.description="Mini Notion-like self-hosted"
```

Et dans `docker-compose.yml`, utiliser le tag versionné :

```yaml
image: ghcr.io/cyril-fl/notes:0.3.0
```

---

### Résumé exécutif

Le projet est **bien architecturé** avec des fondations solides (Nuxt 4, TipTap, MongoDB, Docker, types Zod, optimistic updates). La fiabilité est désormais assurée : auto-save, confirmation de suppression, notifications d'erreur, corbeille avec restauration.

**Les 3 priorités immédiates** pour la prochaine étape :

1. **Drag & drop** — réorganiser notes et dossiers sans supprimer/recréer
2. **Hashtags cliquables** — l'infrastructure existe (API search + filtrage par tags), manque le lien dans l'UI
3. **Breadcrumb de navigation** — le `path` existe dans le modèle, à exploiter pour l'orientation

Version recommandée : **`0.5.0`**, avec un chemin clair vers `1.0.0`.

---

## Tech Stack

- **Frontend** : Nuxt 4, Vue 3, Tailwind CSS 4, Nuxt UI, Reka UI
- **Éditeur** : TipTap 3 (Markdown, TaskList, Link, Highlight, TextAlign)
- **State** : Pinia
- **Backend** : Nitro (Nuxt server), MongoDB (Nitro Storage + connexion directe)
- **Recherche** : MiniSearch (full-text, fuzzy, prefix)
- **Validation** : Zod
- **i18n** : @nuxtjs/i18n (fr-FR)
- **Infra** : Docker, Docker Compose, Node 22-alpine
- **DX** : TypeScript, ESLint, Prettier, Husky, CommitLint

## Setup

```bash
pnpm install
```

## Development

```bash
# Démarrer MongoDB
docker compose -f docker-compose.dev.yml up -d

# Démarrer le serveur de développement sur http://localhost:3000
pnpm dev
```

## Production

```bash
# Build
pnpm build

# Preview local
pnpm preview

# Docker
docker compose up -d
```
