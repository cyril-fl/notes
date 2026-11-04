# notes-go

CLI de prise de notes en Go, avec dossiers, tags et corbeille. Les notes sont en markdown et stockées dans PostgreSQL.

## Prérequis

- Go (version indiquée dans `go.mod`)
- Docker, avec Colima pour `make up` et `make down`
- Un checkout de [plum](https://github.com/cyril-fl/plum) : `go.mod` le remplace par un chemin local (directive `replace`), à adapter à ta machine

## Installation

```sh
cp .env.example .env     # puis renseigne les mots de passe
make up                  # démarre Colima, PostgreSQL et LibreDB Studio
go build -o "$(go env GOPATH)/bin/notes" .
```

Le schéma est créé automatiquement à la première commande.

### Utiliser `notes` depuis n'importe quel dossier

Crée `~/.config/notes/.env` :

```sh
POSTGRES_USER=app
POSTGRES_PASSWORD=...
POSTGRES_DB=notes
POSTGRES_HOST=localhost
PORT=5432
```

- Ce fichier passe avant le `.env` du dossier courant. Sans lui, le `.env` d'un autre projet peut diriger la CLI vers une autre base, où elle créerait son schéma et lancerait sa purge.
- N'y copie pas `CONFIG_FILE=config/config.yaml` : ce chemin est relatif et ne fonctionne que depuis le repo.

## Utilisation

Les commandes `note` et `folder`, ainsi que `move`, `finder` et `graph`, acceptent un UUID ou un chemin :

```sh
notes show note 3f2a...               # par UUID
notes show note /Projets/Go/Idées     # par chemin
```

Un chemin doit correspondre exactement, majuscules comprises. Si plusieurs éléments ont le même chemin, l'erreur liste leurs UUID.

### Notes et dossiers

```sh
notes add folder --title Projets
notes add note --title Idées --content "..." --tag go,sql --parent /Projets
cat idees.md | notes add note --title Idées --content - --parent /Projets

notes show note /Projets/Idées
notes show folder /Projets

notes edit note /Projets/Idées --tag go --untag sql
cat idees.md | notes edit note /Projets/Idées --content -
notes edit folder /Projets --title Travail

notes move /Projets/Idées --parent /Archives
```

Si une valeur manque, un formulaire s'ouvre. Sans terminal (cron, CI), passe tout par les flags, avec `--title` ou `--no-title` pour le titre.

### Parcourir et chercher

```sh
notes finder [dossier]    # explorer, ouvrir, modifier, créer, supprimer
notes graph [note]        # naviguer entre les notes qui partagent des tags
notes search recette      # chercher dans les titres et le contenu
notes recent              # les 20 dernières notes modifiées
```

### Tags

```sh
notes add tag go sql
notes show tag go
notes show tag --list
notes edit tag go golang
notes remove tag sql
```

### Corbeille

```sh
notes remove note /Projets/Idées
notes remove folder /Projets      # avec tout son contenu
notes trash
notes restore <uuid>
```

`restore` n'accepte que des UUID, affichés par `notes trash`. Les éléments sont supprimés définitivement au lancement de la CLI, une fois passé le délai de `trash.retention-days` (30 jours, `0` pour ne jamais purger).

### Scripts

`search`, `recent`, `trash` et `show folder` acceptent `--json`, tout comme `show tag --list`, qui accepte aussi `--yaml` :

```sh
notes search recette --json | jq -r '.[].id'
```

Le détail de chaque commande est dans `notes <commande> --help`.

## API

```sh
make api                                     # dans Docker, avec WEB_ORIGIN défini dans .env
notes serve --origin http://localhost:3000   # ou directement sur la machine
```

`--addr` choisit l'adresse d'écoute (`127.0.0.1:8080` par défaut) et `--origin` l'origine de l'app web autorisée.

| Ressource | Routes (préfixe `/api`) | Corps JSON |
|---|---|---|
| Notes | `POST /notes`, `GET`, `PATCH`, `DELETE /notes/{id}`, `GET /notes/{id}/neighbors` | `title`, `content`, `tags`, `parent_id` (création) |
| Dossiers | `GET`, `POST /folders`, `GET`, `PATCH`, `DELETE /folders/{id}` | `title`, `parent_id` (création) |
| Parcourir | `GET /items?parent=<uuid>` (racine sans `parent`), `POST /move` | `ids`, `parent_id` (racine sans `parent_id`) |
| Tags | `GET`, `POST /tags`, `GET`, `PATCH`, `DELETE /tags/{value}` | `values` (création), `value` (renommage) |
| Graphe | `GET /graph` | |
| Recherche | `GET /search?q=<texte>`, `GET /recent` | |
| Corbeille | `GET`, `DELETE /trash`, `POST /restore` | `ids` (restauration) |

- `PATCH /notes/{id}` ne change que les champs envoyés : sans `tags`, les tags sont conservés, et avec `"tags": []`, ils sont retirés.
- Les erreurs ont la forme `{"error": "..."}`, avec le statut 400, 403, 404, 409, 413 ou 500.
  Seuls une route inconnue (404) et une méthode non gérée (405) répondent en texte brut, car c'est le routeur qui les écrit.
- Les éléments supprimés passent par la corbeille. Le serveur la purge toutes les heures. `DELETE /trash` la vide tout de suite et définitivement.
- `make openapi` génère le contrat OpenAPI dans `docs/swagger.yaml`, à partir des commentaires `@` des routes.

### Sécurité

L'API n'a pas d'authentification : **ne la publie jamais ailleurs que sur `127.0.0.1`**.

- Une requête dont l'en-tête `Origin` diffère de `--origin` est refusée (403). Sans `--origin`, aucune page web ne peut l'appeler.
- Une requête sans `Origin` (`curl`, scripts) est acceptée : tout programme local peut donc appeler l'API.
- L'en-tête `Host` doit valoir `localhost`, `127.0.0.1`, `[::1]` ou l'hôte de `--origin`, ce qui bloque le DNS rebinding.

## Configuration

| Variable | Rôle | Défaut |
|---|---|---|
| `POSTGRES_USER`, `POSTGRES_PASSWORD` | Identifiants de la base | |
| `POSTGRES_DB` | Nom de la base | `notes` |
| `POSTGRES_HOST` | Hôte | `localhost` |
| `PORT` | Port PostgreSQL | `5432` |
| `CONFIG_FILE` | Remplace la configuration embarquée | `config/config.yaml`, intégré au binaire |
| `WEB_ORIGIN` | Origine de l'app web pour `make api` | aucune |

Les commandes, leurs textes d'aide et leurs flags sont définis dans `config/config.yaml`. Ce fichier est intégré à la compilation : recompile après l'avoir modifié, ou pointe `CONFIG_FILE` vers ton fichier.

## Sauvegarde

```sh
make backup    # écrit backups/notes-<date>.sql (ignoré par git)
```

Pour restaurer :

```sh
docker compose exec -T postgres sh -c 'psql -U "$POSTGRES_USER" "$POSTGRES_DB"' < backups/<fichier>.sql
```

Le dump commence par supprimer les tables existantes : la restauration remplace donc tout le contenu de la base.

## Développement

```sh
make setup    # installe les hooks lefthook
make test     # tous les tests, PostgreSQL doit tourner
make test ARGS="./app/shared/services/process/..."
make down     # arrête les conteneurs et Colima
```

LibreDB Studio, une interface pour explorer la base, est sur <http://localhost:3001>.

| Dossier | Contenu |
|---|---|
| `cmd/` | Déclaration des commandes Cobra |
| `app/cli/` | Point d'entrée et affichage de chaque commande |
| `app/shared/services/process/` | Logique métier et requêtes, réutilisables par l'API |
| `app/shared/domain/` | Modèles de données |
| `app/api/` | Routes de l'API, une par dossier |
| `config/` | Configuration et textes des commandes |

## Licence

[MIT](LICENSE)
