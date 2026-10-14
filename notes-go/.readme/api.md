# API Routes — SCN

## Correspondance CLI → API

| CLI | Méthode HTTP | Route API |
|---|---|---|
| `scn note add` | `POST` | `/notes` |
| `scn note show <id-or-path>` | `GET` | `/notes/{id}` ou `/notes/by-path/{path}` |
| `scn note update <id>` | `PATCH` | `/notes/{id}` |
| `scn note move <id> <destination>` | `POST` | `/notes/{id}/move` |
| `scn note remove <id-or-path>` | `DELETE` | `/notes/{id}` ou `/notes/by-path/{path}` |
| `scn folder add <name>` | `POST` | `/folders` |
| `scn folder show <id-or-path>` | `GET` | `/folders/{id}` ou `/folders/by-path/{path}` |
| `scn folder update <id>` | `PATCH` | `/folders/{id}` |
| `scn folder move <id> <destination>` | `POST` | `/folders/{id}/move` |
| `scn folder remove <id-or-path>` | `DELETE` | `/folders/{id}` ou `/folders/by-path/{path}` |
| `scn tag add <value>` | `POST` | `/tags` |
| `scn tag show <id-or-value>` | `GET` | `/tags/{id}` ou `/tags/by-value/{value}` |
| `scn tag update <id>` | `PATCH` | `/tags/{id}` |
| `scn tag remove <id>` | `DELETE` | `/tags/{id}` |
| `scn note show <id> --related` | `GET` | `/notes/{id}/related` |
| `scn finder [query]` | `GET` | `/search` |
| `scn graph` | `GET` | `/graph` |

## Paramètres de requête

### Notes liées

**Route :** `GET /notes/{id}/related`

| Paramètre | Correspondance CLI | Description |
|---|---|---|
| `min_weight` | `--min-weight` | Poids minimum |
| `tag` | Filtre par tag | Limite les résultats à un tag |
| `id` | Filtre par identifiant | Limite les résultats à une note précise |

### Recherche

**Route :** `GET /search`

| Paramètre | Correspondance CLI | Description |
|---|---|---|
| `q` | `finder [query]` | Texte recherché |
| `tag` | `--tag` | Filtre par tag |
| `path` | `--path` | Filtre par chemin |
| `type` | `--type` | Filtre par type de ressource |

### Graphe

**Route :** `GET /graph`

| Paramètre | Correspondance CLI | Description |
|---|---|---|
| `folder` | `--folder` | Limite le graphe à un dossier |
| `scope` | `--direct` | Portée : `direct` ou `recursive` |
| `min_weight` | `--min-weight` | Poids minimum des relations |

### Dossiers

**Route :** `GET /folders/{id}`

| Paramètre | Description |
|---|---|
| `tree=true` | Inclut l'arborescence |
| `recursive=true` | Demande les statistiques récursives |

## Gestion des tags d'une note

| CLI | Méthode HTTP | Route API |
|---|---|---|
| `scn note update <id> --add-tag <tag>` | `POST` | `/notes/{id}/tags` |
| `scn note update <id> --remove-tag <tag>` | `DELETE` | `/notes/{id}/tags/{tagId}` |

## Déplacement des ressources

Les routes suivantes reçoivent la destination dans le corps JSON de la requête :

- `POST /notes/{id}/move`
- `POST /folders/{id}/move`

Exemple :

```json
{
  "destination": "dossier/notes"
}
```

## Conventions API

- **Format :** JSON pour les corps de requête et les réponses.
- **Identifiants :** utilisés dans les routes des ressources.
- **Chemins :** accessibles via des routes dédiées ; les `/` contenus dans les chemins doivent être correctement gérés.
- **Options CLI uniquement :** `--json`, `--interactive` et `--force` ne deviennent pas des routes API.
- **Confirmations :** gérées par la CLI, pas par l'API.