# notes-go

CLI de prise de notes en Go, avec dossiers, tags et corbeille. Les notes sont en markdown et stockées dans PostgreSQL.

## Prérequis

- Go (version indiquée dans `go.mod`)
- Docker, avec Colima pour `make up` et `make down`

## Installation

```sh
cp .env.example .env     # puis renseigne les mots de passe
make up                  # démarre Colima, PostgreSQL, Silo et LibreDB Studio
go build -o "$(go env GOPATH)/bin/notes" .
```

Sans le repo, par exemple sur le serveur : `go install github.com/cyril-fl/notes/notes-go@latest` installe la dernière version de `main` sous le nom `notes-go`, configurée par `~/.config/notes/.env` (ci-dessous).

Le schéma est créé automatiquement à la première commande.

### Utiliser `notes` depuis n'importe quel dossier

Crée `~/.config/notes/.env` :

```sh
POSTGRES_USER=app
POSTGRES_PASSWORD=...
POSTGRES_DB=notes
POSTGRES_HOST=localhost
PORT=5432
# Pour export et import, qui lisent et écrivent les images
S3_ENDPOINT=localhost:9000
S3_ACCESS_KEY=notes
S3_SECRET_KEY=...
S3_BUCKET=notes
```

- Ce fichier passe avant le `.env` du dossier courant. Sans lui, le `.env` d'un autre projet peut diriger la CLI vers une autre base, où elle créerait son schéma et lancerait sa purge.
- N'y copie pas `CONFIG_FILE=config/config.yaml` : ce chemin est relatif et ne fonctionne que depuis le repo.

## Utilisation

### Comptes

Chaque note, dossier et tag appartient à un compte, et un compte ne voit jamais ceux des autres. Les comptes se créent sur le serveur, par la CLI :

```sh
notes user add cyril            # demande le mot de passe deux fois (8 caractères au moins)
notes login                     # nom et mot de passe, puis toutes les commandes agissent pour ce compte
notes logout
notes reset-password cyril      # mot de passe oublié : nouveau mot de passe, tous les appareils déconnectés
notes rotate-secret             # rechiffre les clés des comptes par un nouveau NOTES_SECRET (voir Chiffrement)
notes account purge             # vide le compte connecté (notes, dossiers, tags, images), après avoir tapé son nom
```

- **Scripts :** `--password-stdin` lit le mot de passe sur l'entrée standard, pour `user add` et `reset-password`.
- **Session de la CLI :** elle est gardée dans `~/.config/notes/session` sous Linux et `~/Library/Application Support/notes/session` sous macOS, lisible par l'utilisateur seul. Elle dure 30 jours, prolongés à chaque usage.

Les commandes `note` et `folder`, ainsi que `move`, `finder` et `graph`, acceptent un UUID ou un chemin :

```sh
notes show note 3f2a...               # par UUID
notes show note /Projets/Go/Idées     # par chemin
```

Un chemin doit correspondre exactement, majuscules comprises. Si plusieurs éléments ont le même chemin, l'erreur liste leurs UUID.

### Notes et dossiers

```sh
notes add folder --title Projets
notes add note --content "# Idées" --tag go,sql --parent /Projets
cat idees.md | notes add note --content - --parent /Projets

notes show note /Projets/Idées
notes show folder /Projets

notes edit note /Projets/Idées --tag go --untag sql
notes edit note /Projets/Idées --pin       # --unpin pour retirer
notes edit note /Projets/Idées --lock      # lecture seule, --unlock pour retirer
cat idees.md | notes edit note /Projets/Idées --content -
notes edit folder /Projets --title Travail
notes edit folder /Projets --pin            # en favori, --unpin pour retirer

notes move /Projets/Idées --parent /Archives
```

Si une valeur manque, un formulaire s'ouvre. Sans terminal (cron, CI), passe tout par les flags.

Deux dossiers d'un même dossier ne peuvent pas porter le même nom, majuscules comprises : créer, renommer, déplacer ou restaurer un dossier à côté d'un homonyme est refusé (409 dans l'API). Une base créée avant cette règle voit ses doublons renommés au démarrage, `Nom (2)`, `Nom (3)`… Les notes, elles, peuvent partager un titre.

Une note n'a pas de titre à part : c'est son premier titre markdown (`#`, `##`…, quel que soit le niveau), sinon les 50 premiers caractères de sa première ligne de texte, sans la mise en forme, le code ni les lignes faites seulement de `#tags`. Une note sans texte reste sans titre. Le titre suit le contenu à chaque modification, et le chemin d'une note (`/Projets/Idées`) avec lui.

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

Les tags d'une note sont les `#mots` de son texte, code et liens exclus : `#idée` dans le contenu crée le tag et le lie à la note, et l'effacer l'en retire. `--tag` écrit `#tag` à la fin du texte et `--untag` l'efface. Renommer ou supprimer un tag corrige le texte de toutes ses notes, et une note qui se retrouve vide part à la corbeille. Un tag sans plus aucune note disparaît. Un tag contient des lettres, chiffres, `_` ou `-`, avec au moins une lettre.

### Liens entre notes

Écris `[[Titre]]` dans une note pour la relier à la note de ce titre, majuscules ou non, hors code. À l'enregistrement, un `[[Titre]]` qui désigne une seule note devient `[Titre](note:<uuid>)` : le lien suit désormais cette note par son id, même si une autre prend son titre, et le texte entre crochets se met à jour quand elle change de titre. Si aucune note, ou plusieurs, ont ce titre, `[[Titre]]` attend : il sera relié dès qu'une seule note portera ce titre. Pour départager, écris son chemin, `[[/Projets/Idées]]`.

`[[Titre#Section]]` vise un titre de la note : il devient `[Titre#Section](note:<uuid>#Section)`, garde la section quand la note change de titre, et redevient `[[Titre#Section]]` à l'export. Dans l'app, taper `#` après le nom d'une note propose ses titres, et un clic ouvre la note à ce titre. `[[#Section]]` vise un titre de la note elle-même : il reste tel quel, `[[` puis `#` propose les titres de la note, et un clic y fait défiler.

`![[Titre]]`, comme dans Obsidian, inclut la note : il devient `![Titre](note:<uuid>)` comme un lien, et redevient `![[Titre]]` à l'export. Dans l'app, la note incluse s'affiche en lecture sous un lien qui l'ouvre ; ses propres inclusions restent des liens, pour qu'une note qui s'inclut elle-même, ou qui inclut une note qui l'inclut, s'arrête là.

`GET /api/notes/{id}/neighbors` donne ses liens (`links`) et ses rétroliens (`backlinks`), et le graphe relie aussi les notes liées. Dans l'app, taper `[[` propose les notes, et un clic sur un lien l'ouvre. `GET /api/links/broken` liste les liens des notes vivantes qui ne mènent nulle part : `unresolved` (aucune autre note de ce titre ou chemin, ou un lien `#Section` vers un titre que la note n'a pas), `ambiguous` (plusieurs) `missing` (lien vers une note mise à la corbeille ou purgée) ou `file` (image ou pièce jointe `/api/files/<id>` que le compte n'a pas : importée d'un autre compte, purgée).

### Corbeille

```sh
notes remove note /Projets/Idées
notes remove folder /Projets      # avec tout son contenu
notes trash
notes restore <uuid>
```

`restore` n'accepte que des UUID, affichés par `notes trash`. Les éléments sont supprimés définitivement au lancement de la CLI, une fois passé le délai de conservation du compte (30 jours par défaut, réglable dans Réglages > Corbeille de l'app, « Jamais » pour ne pas purger).

### Export et import

```sh
notes export ~/notes-md              # toutes les notes
notes export ~/projets /Projets      # un dossier
notes import ~/notes-md --parent /Archive
notes import ~/Obsidian/vault        # à la racine
```

`export` écrit dans un dossier nouveau ou vide, jamais par-dessus des fichiers : un dossier devient un répertoire, une note un fichier `<titre>.md` avec son texte tel quel (titre et `#tags` compris). Les liens `[Titre](note:<uuid>)` redeviennent `[[Titre]]`, puisque les ids n'ont pas de sens hors de cette base. Deux notes du même titre dans le même dossier donnent `Titre.md` et `Titre (2).md`.

`import` fait l'inverse, tout ou rien : un nom de dossier déjà pris au même endroit arrête tout l'import, sans rien laisser derrière. Les fichiers cachés (`.obsidian`, `.git`…) et ceux qui ne sont pas des `.md` sont ignorés. Un texte sans titre markdown reçoit le nom de son fichier en tête, `# Nom`, comme dans Obsidian. Les `[[liens]]` se relient à l'import, comme à la frappe.

Les images suivent les notes. À l'export, celles qu'une note cite sont copiées dans `_files/` à la racine de l'export, et la note les cite par un chemin relatif (`../_files/<uuid>.png`). À l'import, une image citée par un chemin relatif (`![](photo.png)`, `![](<photo de mer.png>)`) devient une image de la note, où qu'elle soit dans le dossier. `_files/` ne devient pas un dossier. Une image absente ou qui n'est pas en PNG, JPEG, GIF ou WebP garde son lien tel quel. Un lien vers un fichier du dossier (`[rapport](rapport.docx)`) ou un `![[son.mp3]]` qui n'est pas une image en fait une pièce jointe, citée par un lien ; un lien vers une note ou un type refusé reste tel quel. À l'export, un zip garde l'extension de son nom (`.docx`, `.xlsx`…). Les `![[image.png]]` et `![[image.png|300]]` d'Obsidian aussi : l'image est cherchée par son nom dans tout le dossier (celle du dossier de la note d'abord), leur largeur devient la taille la plus proche de l'éditeur (`"small"` jusqu'à 300 px, `"medium"` jusqu'à 600, rien au-delà).

### Images

Coller ou glisser une image dans l'éditeur de l'app l'envoie à l'API, qui la range dans [Silo](https://github.com/pgsty/silo) (un stockage S3, fork de MinIO, démarré par `make up`) et la décrit dans la table `files`. La note la cite en markdown : `![nom](/api/files/<uuid>)`. Un autre fichier (PDF, zip, donc aussi les documents Office, MP3, WAV, Ogg, MP4, WebM ; 20 Mo au plus) devient une pièce jointe, citée par un lien `[nom](/api/files/<uuid>)` : l'API l'envoie en téléchargement sous son nom, il suit les mêmes règles de purge et d'export qu'une image. L'avatar reste une image. `GET /api/files` liste les fichiers du compte (nom, type, taille), les plus lourds d'abord, sans ceux d'une note purgée ou de données effacées.

- PNG, JPEG, GIF et WebP seulement (pas de SVG, qui peut contenir des scripts), 20 Mo au plus. Le type est lu dans le fichier, pas dans ce qu'envoie le navigateur.
- Une image suit sa note dans la corbeille : quand la note est effacée pour de bon (délai écoulé, corbeille vidée, ou son dossier purgé), ses images partent avec elle à la purge horaire suivante, sauf si une autre note les cite encore.
- Une image retirée du texte d'une note attend le délai de corbeille du compte, compté depuis qu'elle n'est plus citée (jamais si la corbeille garde tout) : d'ici là, l'annulation ou le lien recollé la font revenir.
- Le navigateur ne parle jamais à Silo : il passe par `GET /api/files/{id}`, protégé par la session. La console de Silo est sur http://localhost:9001 (identifiants `S3_ACCESS_KEY` / `S3_SECRET_KEY`).

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

`--addr` choisit l'adresse d'écoute (`127.0.0.1:8080` par défaut) et `--origin` l'origine de l'app web autorisée, `--web` un dossier de l'app générée (`bun run generate`) servi à la même adresse que l'API. Quand l'API n'est joignable que depuis la machine (`127.0.0.1`), elle prend l'adresse de chaque appareil dans le `X-Forwarded-For` du proxy (`tailscale serve`), pour que le délai après de mauvais mots de passe ne bloque que l'appareil fautif. Dans un conteneur, `--published` donne l'adresse publiée côté machine (`NOTES_LISTEN`), à la place de `--addr`.

| Ressource | Routes (préfixe `/api`) | Corps JSON |
|---|---|---|
| Compte | `POST /login`, `POST /logout`, `GET`, `PATCH /me`, `PATCH /me/password` | `name`, `password` (connexion) ; réglages : `locale`, `theme`, `color_mode`, `view`, `trash_retention_days` (0 à 3650, 0 pour ne jamais purger), `name`, `avatar` (id d'une image du compte, `""` pour aucune) ; mot de passe : `current`, `new` |
| Notes | `POST /notes`, `GET`, `PATCH`, `DELETE /notes/{id}`, `GET /notes/{id}/neighbors`, `GET /pinned` | `content`, `tags`, `parent_id` (création), `pinned`, `readonly` |
| Dossiers | `GET`, `POST /folders`, `GET`, `PATCH`, `DELETE /folders/{id}` | `title`, `parent_id` (création), `pinned` |
| Parcourir | `GET /items?parent=<uuid>` (racine sans `parent`), `POST /move` | `ids`, `parent_id` (racine sans `parent_id`) |
| Tags | `GET`, `POST /tags`, `GET`, `PATCH`, `DELETE /tags/{value}` | `values` (création), `value` (renommage) |
| Graphe | `GET /graph` (tags comptés), `GET /graph/notes` (notes et tags qui les relient) | |
| Recherche | `GET /search?q=<texte>`, `GET /recent` | |
| Corbeille | `GET`, `DELETE /trash` (tout), `DELETE /trash/{id}` (une suppression et ce qui est parti avec), `POST /restore` | `ids` (restauration) |
| Export / import | `GET /export` (zip), `POST /import?parent=<uuid>` (zip en corps, `Content-Type: application/zip`, ou formulaire `multipart/form-data` de fichiers `files` nommés par leur chemin, `dossier/note.md` ; racine sans `parent`) | |
| Images et pièces jointes | `POST /files?name=<nom>` (le fichier en corps), `GET /files/{id}` | |

- Tout, sauf `POST /login`, demande une session, sinon 401. La connexion pose un cookie `notes_session` (`HttpOnly`, `Secure`, `SameSite=Strict`), qui dure 30 jours prolongés à chaque requête. Les scripts peuvent envoyer le jeton dans `Authorization: Bearer <jeton>`. Les sessions expirées sont effacées de la base à la purge horaire du serveur.
- Après des mauvais mots de passe d'affilée (par adresse pour `POST /login`, par compte pour `PATCH /me/password`), la route répond 429 avec `Retry-After` (en secondes) : rien jusqu'à 3 échecs, puis 30 s, 1 min, 5 min, 15 min, et 1 h dès 8. Une réussite remet le compteur à zéro.
- Le titre d'une note vient de son contenu, comme dans la CLI : l'API refuse un champ `title` pour une note.
- `PATCH /notes/{id}` ne change que les champs envoyés. `{"pinned": true}` seul épingle la note sans toucher à son texte ni à sa date de modification. Sans `tags`, les tags sont les `#mots` du contenu. Avec `tags`, cette liste réécrit le contenu : `"tags": []` efface tous ses `#mots`. Il répond la note complète, comme `GET` (texte enregistré, tags, chemin). `{"readonly": true}` verrouille la note : `content` et `tags` sont alors refusés (409), sauf avec `"readonly": false` dans la même requête. Avec `updated_at` (celui lu par `GET`, ou `created_at` pour une note jamais modifiée), `content` et `tags` sont refusés (409) si la note a changé depuis : un appareil resté ouvert n'écrase pas ce qu'un autre a écrit.
- `GET /notes/{id}/versions` liste les versions de la note, de la plus récente à la plus ancienne, sans leur texte (`id`, `conflict`, `created_at`) ; `GET /notes/{id}/versions/{version}` en lit une avec son `content`. Un `PATCH` (ou `notes edit note`) qui change le texte garde l'ancien à la première sauvegarde d'une session d'écriture (la note non modifiée depuis 15 minutes), et au moins une fois par heure dans une longue session ; une sauvegarde sans changement ne compte pas. Les versions ne sont pas limitées en nombre. Restaurer une version, c'est un `PATCH` de son `content` ; les versions partent avec la note quand elle est supprimée pour de bon.
- `POST /notes/{id}/versions` avec `{"content": "..."}` garde un texte refusé par un 409 comme une version marquée `conflict` (201), pour ne pas le perdre ; rien si c'est déjà le texte de la note (204).
- `PATCH /folders/{id}` avec `{"pinned": true}` met le dossier en favori. `GET /pinned` liste les notes épinglées et les dossiers favoris, avec leur `kind`.
- Les erreurs ont la forme `{"error": "..."}`, avec le statut 400, 401, 403, 404, 409, 413, 415, 429 ou 500.
  Seuls une route inconnue (404) et une méthode non gérée (405) répondent en texte brut, car c'est le routeur qui les écrit.
- `PATCH /me` avec `name` renomme le compte (409 si le nom est pris). `PATCH /me/password` vérifie `current`, puis déconnecte les autres appareils : celui qui l'envoie reste connecté. L'image de `avatar` n'est pas purgée tant qu'elle sert d'avatar.
- `DELETE /me/data` efface pour de bon toutes les notes, dossiers, tags et images du compte, corbeille comprise. Le compte, ses réglages et ses sessions restent. C'est la « Zone de danger » des réglages de l'app web.
- Les éléments supprimés passent par la corbeille. Le serveur la purge toutes les heures. `DELETE /trash` la vide tout de suite et définitivement.
- `GET /export` et `POST /import` font comme `notes export` et `notes import`, en zip ; l'import prend aussi un formulaire de fichiers, leurs chemins donnant les dossiers. L'import va dans le dossier `parent`, à la racine sans lui. Il répond `{"folders": n, "notes": n}`, ou 409 si un dossier du zip existe déjà à cet endroit, 400 si deux fichiers du formulaire ont le même chemin.
- `POST /files` répond le fichier créé (`id`, `name`, `content_type`, `size`), 415 si ce n'est pas une image acceptée. `GET /files/{id}` renvoie l'image, mise en cache par le navigateur : une image ne change jamais.
- `make openapi` génère le contrat OpenAPI dans `docs/swagger.yaml`, à partir des commentaires `@` des routes.

### Sécurité

Chaque requête doit venir d'un compte connecté. Garde quand même l'API sur `127.0.0.1` derrière Tailscale ou un proxy HTTPS : le mot de passe et le cookie ne doivent jamais passer en clair sur un réseau.

- Une requête dont l'en-tête `Origin` diffère de `--origin` est refusée (403). Sans `--origin`, aucune page web ne peut l'appeler. Le cookie `SameSite=Strict` ajoute une protection contre les autres sites.
- Une requête sans `Origin` (`curl`, scripts) est acceptée si elle porte une session.
- L'en-tête `Host` doit valoir `localhost`, `127.0.0.1`, `[::1]` ou l'hôte de `--origin`, ce qui bloque le DNS rebinding.

### Chiffrement

Le texte et les titres des notes, les noms des dossiers, les tags, les snippets et les noms des images sont chiffrés dans la base, et les images dans Silo (AES-256-GCM). Chaque compte a sa propre clé, tirée au hasard à sa première connexion et rangée dans `users.data_key`, elle-même chiffrée par le secret du serveur `NOTES_SECRET`. Une base, une sauvegarde ou un disque volés restent illisibles sans ce secret. Un serveur compromis, lui, le détient : le chiffrement ne protège pas contre ça.

- Génère la clé avec `openssl rand -base64 32` et mets la même dans `.env` (CLI) et `.env.docker` (API).
- **Garde une copie de la clé dans un gestionnaire de mots de passe, jamais dans une sauvegarde.** Si tu la perds, les notes deviennent illisibles, même avec une sauvegarde.
- Un mot de passe oublié ne fait rien perdre : `notes reset-password` suffit, car les clés ne dépendent pas du mot de passe.
- Changer le secret : `make backup`, mets le nouveau dans `NOTES_SECRET` (`.env`), puis `NOTES_OLD_SECRET=<ancien> notes rotate-secret`. Seules les clés des comptes sont rechiffrées, pas les notes. Mets ensuite le nouveau secret dans `.env.docker` et lance `make api`. Un second passage ne fait rien.
- `notes export` sort du texte en clair.
- Pour retrouver un tag, un dossier ou une note par son nom, et garder ces noms uniques, la base garde une empreinte HMAC du nom en minuscules (`title_hash`, `value_hash`). Elle ne se lit pas, mais deux noms identiques ont la même empreinte. Les chemins, les tris et la recherche se font dans l'API, après déchiffrement.
- Ne sont pas chiffrés : les dates et l'arborescence (quel élément est dans quel dossier).

## Configuration

| Variable | Rôle | Défaut |
|---|---|---|
| `POSTGRES_USER`, `POSTGRES_PASSWORD` | Identifiants de la base | |
| `POSTGRES_DB` | Nom de la base | `notes` |
| `POSTGRES_HOST` | Hôte | `localhost` |
| `PORT` | Port PostgreSQL | `5432` |
| `CONFIG_FILE` | Remplace la configuration embarquée | `config/config.yaml`, intégré au binaire |
| `WEB_ORIGIN` | Origine de l'app web pour `make api` | aucune |
| `S3_ENDPOINT` | Adresse de Silo, sans `http://` (`note-silo:9000` dans `.env.docker`) | |
| `S3_ACCESS_KEY`, `S3_SECRET_KEY` | Identifiants de Silo, qui le créent aussi au premier `make up` | |
| `S3_BUCKET` | Bucket des images, créé au premier lancement | |
| `S3_SECURE` | `true` pour parler à Silo en HTTPS | non |
| `NOTES_SECRET` | Clé de chiffrement des notes, 32 octets en base64 (voir Chiffrement) | aucune : obligatoire pour écrire |

Les commandes, leurs textes d'aide et leurs flags sont définis dans `config/config.yaml`. Ce fichier est intégré à la compilation : recompile après l'avoir modifié, ou pointe `CONFIG_FILE` vers ton fichier.

## Sauvegarde

```sh
make backup    # écrit backups/notes-<date>.sql et les images dans backups/notes-<date>.files/ (ignorés par git)
```

Pour restaurer :

```sh
docker compose exec -T postgres sh -c 'psql -U "$POSTGRES_USER" "$POSTGRES_DB"' < backups/<fichier>.sql
docker compose cp backups/<fichier>.files silo:/tmp/restore
docker compose exec silo sh -c 'mcli alias set local http://127.0.0.1:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD" && mcli mirror /tmp/restore local/notes'
```

Pour tout effacer, tous comptes confondus, après une sauvegarde : `make purge` (il faut taper `purge` pour confirmer). Les comptes et leurs réglages restent.

Le dump commence par supprimer les tables existantes : la restauration remplace donc tout le contenu de la base. Les images s'ajoutent à celles de Silo, sans rien effacer.

## Développement

```sh
make setup    # installe les hooks lefthook
make test     # tous les tests, PostgreSQL doit tourner
make test ARGS="./app/shared/services/process/..."
make down     # arrête les conteneurs et Colima
```

Les tests tournent dans une base à part, `<POSTGRES_DB>_test`, créée au premier `make test`, chacun dans une transaction annulée. Ils créent le schéma : une migration en cours d'écriture n'atteint donc jamais la base que sert l'API.

LibreDB Studio, une interface pour explorer la base, est sur <http://localhost:3001>. En ligne de commande : `make pgsql` ouvre psql, `make pgsql QUERY="SELECT ..."` lance une requête et s'arrête à la première erreur.

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
