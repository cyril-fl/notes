# Notes
Projet de prise de notes un peu manière obsiediennes.

## Objectif : 
- Faire une cli
- Faire une web app

## Choix de la stack

### Core: Go
- Faire a la fois un CLI et un Back end qui tourne.

### Webapp: Nuxt 

#### Avantages : 

- Je connais déjà
- C'est simplement configurable avec front / back.
- Y a plein de plugins a mettre pour avoir un truc fonctionnel vite

### Data: sqlite + Go Bun ORM

#### Avantages : 

- C'était vachement compliqué a gérer sur la version précédente avec mongo. 
- Bun ORM est natif a go et permet de coder en directement mes interface.
- Une DB sql avec un ORM sera surement plus maintenable avec la structure de données soihaité

## Serveur

Chaque tag `v*` publie l'image `ghcr.io/cyril-fl/notes` (API et app web sur la même origine, amd64 et arm64). Sur le serveur, à côté de `compose.prod.yaml` et d'un `.env` copié de `.env.prod.example` :

```sh
docker compose -f compose.prod.yaml up -d                                    # Postgres, Silo, Notes et LibreDB Studio
echo '<mot de passe>' | docker exec -i note-api /notes user add <nom> --password-stdin
docker compose -f compose.prod.yaml pull && docker compose -f compose.prod.yaml up -d   # mise à jour
```

### Comptes

Le premier compte de la base est admin : lui seul invite, liste et supprime les autres, depuis les réglages de l'app (section *Comptes*). Il se crée de trois façons :

- la base est vide, on ouvre l'app : elle mène droit à la page de création de compte ;
- `NOTES_HOST_NAME` et `NOTES_HOST_PASSWORD` dans le `.env` : Notes le crée au démarrage, puis ces deux lignes ne servent plus (à vider) ;
- la commande `notes user add` ci-dessus, qui marche toujours.

Ensuite, `NOTES_ADD_USER_MODE` décide qui peut créer un compte :

- `invitation` (par défaut) : l'admin crée un lien d'invitation dans les réglages et l'envoie. Le lien sert une fois, expire au bout de 7 jours, et c'est l'invité qui choisit son nom et son mot de passe — il ne passe jamais par l'admin. Ailleurs, la page de création de compte est fermée et le lien vers elle n'apparaît pas sur la page de connexion.
- `public` : n'importe qui atteignant l'app crée son compte. À réserver à un réseau de confiance (Tailscale) : rien ne limite le nombre de créations.

Supprimer un compte depuis le panneau efface pour de bon ses notes, dossiers, tags et images. Un admin ne peut ni se supprimer lui-même, ni supprimer le dernier admin.

Notes écoute en HTTP sur `NOTES_LISTEN` (`127.0.0.1:8080` par défaut) : le cookie de session exige HTTPS, un proxy devant le fournit. `WEB_ORIGIN` est l'adresse que tape le navigateur. LibreDB Studio, l'interface de la base, écoute sur `LIBREDB_LISTEN` (`127.0.0.1:3001`) : on s'y connecte à `note-postgres:5432` avec l'utilisateur `POSTGRES_USER`.

Au démarrage, Notes ajoute à la base les colonnes des nouvelles versions. Un renommage ou un changement de type reste à faire à la main, indiqué dans le message du commit.

## Qualité

- `make setup` dans `notes-go/` installe les hooks git (lefthook) : gofmt, Prettier et actionlint (les workflows) avant un commit, et le préfixe du message (`feat:`, `fix:`…) vérifié.
- La CI GitHub Actions lance à chaque push les tests Go, le lint, le format, les tests, le typecheck et le build de l'app, vérifie que `swagger.yaml` et `api.d.ts` sont à jour, puis les e2e sur la stack Docker.
