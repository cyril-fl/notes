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
| **Organisation** | Dossiers hiérarchiques (arbre récursif)                              | Fonctionnel |
| **Organisation** | Création de dossier dans un dossier                                  | Fonctionnel |
| **Organisation** | Création de note dans un dossier                                     | Fonctionnel |
| **Organisation** | Renommage de dossier (inline)                                        | Fonctionnel |
| **Organisation** | Suppression (note, dossier + descendants)                            | Fonctionnel |
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
| **Infra**        | Docker multi-stage + Docker Compose                                  | Fonctionnel |
| **Infra**        | MongoDB persistant                                                   | Fonctionnel |
| **DX**           | ESLint, Prettier, Husky, CommitLint                                  | Fonctionnel |

---

### 2. Fonctionnalités essentielles manquantes

#### CORE INDISPENSABLE (pré-requis pour un usage quotidien)

| #   | Fonctionnalité                              | Justification                                                                                              | Complexité |
| --- | ------------------------------------------- | ---------------------------------------------------------------------------------------------------------- | ---------- |
| C1  | **Recherche globale**                       | Impossible de retrouver une note sans parcourir l'arbre. Bloquant dès 20+ notes                            | Moyenne    |
| C2  | **Auto-save fiable avec indicateur**        | Pas de feedback visuel de sauvegarde. L'utilisateur ne sait pas si ses changements sont persistés          | Faible     |
| C3  | **Confirmation avant suppression**          | La suppression d'un dossier détruit tous les descendants sans confirmation — perte de données irréversible | Faible     |
| C4  | **Drag & drop pour réorganiser**            | Déplacer une note/dossier nécessite actuellement de supprimer et recréer. Essentiel pour l'organisation    | Haute      |
| C5  | **Page 404 / gestion des routes invalides** | Navigation vers un ID inexistant = page blanche                                                            | Faible     |
| C6  | **Gestion d'erreurs utilisateur**           | Pas de toast/notification en cas d'échec API. L'utilisateur ne voit pas les erreurs                        | Faible     |

#### AMÉLIORATION UX (qualité perçue, rétention)

| #   | Fonctionnalité                                   | Justification                                                                              | Complexité |
| --- | ------------------------------------------------ | ------------------------------------------------------------------------------------------ | ---------- |
| U1  | **Recherche par hashtag** (clic → filtre)        | Les hashtags sont affichés mais non cliquables/filtrables                                  | Faible     |
| U2  | **Tri et filtres** (date, titre, récents)        | Aucun moyen de trier les notes dans un dossier                                             | Faible     |
| U3  | **Note épinglée / favoris**                      | Pattern standard pour accès rapide aux notes fréquentes                                    | Moyenne    |
| U4  | **Breadcrumb de navigation**                     | Le `path` existe dans le modèle mais n'est pas affiché. L'utilisateur se perd dans l'arbre | Faible     |
| U5  | **Raccourcis clavier globaux**                   | Cmd+N (nouvelle note), Cmd+K (recherche), Cmd+S (save explicite)                           | Moyenne    |
| U6  | **Compteur de mots/caractères**                  | Attendu dans un éditeur de texte                                                           | Faible     |
| U7  | **Responsive / Mobile**                          | Pas d'adaptation mobile visible — sidebar fixe                                             | Moyenne    |
| U8  | **Persistance de l'état sidebar** (localStorage) | La largeur du panel est perdue au rechargement                                             | Faible     |

#### FONCTIONNALITÉS AVANCÉES (différenciation, v1+)

| #   | Fonctionnalité                       | Justification                                                                  | Complexité |
| --- | ------------------------------------ | ------------------------------------------------------------------------------ | ---------- |
| A1  | **Authentification basique**         | Aucune protection — critique pour le self-host exposé sur Internet             | Haute      |
| A2  | **Graph view** (liens entre notes)   | Route `/graph` existe, libs installées (graphology/sigma), mais non implémenté | Haute      |
| A3  | **Images / upload de fichiers**      | Pas de support média dans l'éditeur                                            | Haute      |
| A4  | **Export** (Markdown, PDF, HTML)     | Aucun moyen d'extraire ses données                                             | Moyenne    |
| A5  | **Import** (Markdown, .md files)     | Aucun moyen d'importer du contenu existant                                     | Moyenne    |
| A6  | **Corbeille / soft delete**          | La suppression est définitive, pas de récupération possible                    | Moyenne    |
| A7  | **Historique / versioning de notes** | Pas de possibilité de revenir en arrière sur le contenu                        | Haute      |
| A8  | **Multi-langue** (en + fr)           | Seul le français est supporté, limite l'adoption                               | Faible     |
| A9  | **PWA / offline**                    | Pas de service worker, inutilisable hors connexion                             | Haute      |
| A10 | **API de santé** (`/api/health`)     | Nécessaire pour monitoring Docker/orchestration                                | Faible     |

---

### 3. Hypothèses formulées

- **Pas d'auto-save explicite détecté** : un `watch()` existe sur le contenu de l'éditeur mais sans debounce visible ni indicateur UI — le mécanisme existe partiellement mais manque de fiabilité perçue.
- **Graph view** : les dépendances (graphology, sigma) sont installées et la route existe, mais aucun composant ne les utilise réellement → placeholder.
- **Le delete cascade** sur les dossiers semble fonctionnel (`getRelatedIds`) mais sans modal de confirmation dans l'UI.
- **Pas de tests** : aucun fichier de test détecté (ni unitaire, ni e2e). Manque significatif mais non bloquant pour le versioning initial.

---

### 4. Estimation de version actuelle

| Critère                 | Évaluation                            |
| ----------------------- | ------------------------------------- |
| Structure de l'app      | Solide, bien architecturée            |
| CRUD complet            | Oui                                   |
| Éditeur fonctionnel     | Oui                                   |
| Navigation hiérarchique | Oui                                   |
| Recherche               | Non                                   |
| Auth                    | Non                                   |
| Tests                   | Non                                   |
| Stabilité API           | En cours de définition                |
| Production-ready        | Non (pas d'auth, pas de health check) |

#### Version recommandée : `0.3.0`

**Justification :**

- `0.x.x` → Pas encore stable / pas de contrat d'API public
- `0.1.0` serait trop bas — le CRUD, l'éditeur, la hiérarchie de dossiers et le Docker sont fonctionnels
- `0.3.0` reflète un projet avec des fondations solides et 3 "milestones" implicites déjà franchies : (1) structure initiale, (2) éditeur + CRUD, (3) navigation + dossiers + Docker
- `1.0.0` sera atteint quand : recherche + auth + auto-save fiable + tests minimaux seront en place

---

### 5. Stratégie de versioning (SemVer)

```
MAJOR.MINOR.PATCH  →  0.3.0
```

#### Roadmap suggérée

| Version | Milestone                 | Contenu clé                                                                              |
| ------- | ------------------------- | ---------------------------------------------------------------------------------------- |
| `0.3.0` | **Actuel**                | CRUD, éditeur TipTap, dossiers, Docker                                                   |
| `0.4.0` | **Recherche & Fiabilité** | Recherche globale, auto-save indicator, confirmation delete, toasts d'erreur, breadcrumb |
| `0.5.0` | **Organisation**          | Drag & drop, tri/filtres, favoris, hashtags cliquables, raccourcis clavier               |
| `0.6.0` | **Self-host ready**       | Auth basique (password ou token), health check API, responsive mobile                    |
| `0.7.0` | **Contenu riche**         | Upload images, export MD/PDF, import MD                                                  |
| `0.8.0` | **Avancé**                | Graph view, corbeille/soft delete, multi-langue                                          |
| `0.9.0` | **Qualité**               | Tests unitaires + e2e, PWA/offline, historique de notes                                  |
| `1.0.0` | **Stable**                | API stabilisée, documentation, première release publique                                 |

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
  "version": "0.3.0"
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
{ "status": "ok", "version": "0.3.0", "uptime": 12345 }
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

Le projet est **bien architecturé** avec des fondations solides (Nuxt 4, TipTap, MongoDB, Docker, types Zod, optimistic updates). Il est clairement au-dessus d'un simple prototype mais en-dessous d'un produit utilisable au quotidien.

**Les 3 priorités immédiates** pour débloquer l'usage réel :

1. **Recherche globale** — sans ça, l'app ne scale pas au-delà de 20 notes
2. **Feedback de sauvegarde** — confiance utilisateur
3. **Confirmation de suppression** — sécurité des données

Version recommandée : **`0.3.0`**, avec un chemin clair vers `1.0.0`.

---

## Tech Stack

- **Frontend** : Nuxt 4, Vue 3, Tailwind CSS 4, Nuxt UI, Reka UI
- **Éditeur** : TipTap 3 (Markdown, TaskList, Link, Highlight, TextAlign)
- **State** : Pinia
- **Backend** : Nitro (Nuxt server), MongoDB via Nitro Storage
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
