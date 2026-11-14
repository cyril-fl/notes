# MCP : donner les notes à un agent

`notes mcp` est la troisième façade sur les notes, à côté de la CLI et de l'API : un serveur MCP qui
donne à un agent (Claude Code, Claude Desktop, tout client MCP) des outils pour chercher, lire,
écrire, ranger et supprimer les notes d'un compte, avec leurs fichiers, leur historique et leurs liens.

Ce n'est pas une API HTTP : il n'écoute sur aucun port, n'a pas de routes, ne vérifie aucun jeton
dans une requête. Les trois points à comprendre sont donc ailleurs : **comment l'agent le lance**,
**pour quel compte il agit**, et **ce qu'il peut faire**.

## 1. Comment l'agent le lance

Le transport est **stdio** : le client MCP démarre `notes mcp` comme sous-processus et échange du
JSON-RPC sur son entrée et sa sortie standard. Rien d'autre ne doit être écrit sur stdout, ce qui
casserait le dialogue.

Conséquence : **on ne lance pas `notes mcp` à la main**. Tapé dans un terminal, il attend du JSON et
ne répond rien d'utile. C'est le client qui le démarre, l'arrête et parle avec lui. Il faut donc
seulement le déclarer chez le client, et le faire tourner là où la base est joignable.

Avec Claude Code, le binaire installé (`go install github.com/cyril-fl/notes/notes-go@latest`) et
l'environnement de la CLI déjà en place :

```sh
claude mcp add notes -- notes mcp
```

Dans un `.mcp.json` ou dans la configuration de Claude Desktop, l'environnement doit être donné
explicitement : une application lancée depuis le Finder n'hérite pas du shell.

```json
{
  "mcpServers": {
    "notes": {
      "command": "notes",
      "args": ["mcp"],
      "env": {
        "CONFIG_FILE": "/opt/docker/notes/config/config.yaml",
        "POSTGRES_HOST": "localhost",
        "POSTGRES_PORT": "5432",
        "POSTGRES_DB": "notes",
        "POSTGRES_USER": "app",
        "POSTGRES_PASSWORD": "…",
        "NOTES_SECRET": "…"
      }
    }
  }
}
```

Il lui faut exactement ce qu'il faut à la CLI : les variables Postgres, `NOTES_SECRET` (sans quoi les
notes restent illisibles) et `CONFIG_FILE`. Comme toute commande de la CLI, il crée au démarrage les
tables ou colonnes manquantes et vide la corbeille arrivée à échéance.

Un fichier de configuration d'un client MCP n'est pas un endroit pour `NOTES_SECRET` ni pour le mot
de passe Postgres : sur le serveur, préférer l'environnement du processus de l'agent. Le plus simple
reste d'ailleurs de faire tourner l'agent sur le serveur, où la base est locale et `notes` déjà
installé : il n'y a alors rien de plus à installer.

## 2. Pour quel compte il agit

Il n'y a **ni port, ni jeton d'API, ni mot de passe à passer**. `notes mcp` parle à Postgres comme le
reste de la CLI, et agit pour le compte de la **session CLI de la machine** — le jeton écrit par
`notes login` dans `~/Library/Application Support/notes/session` (macOS) ou
`~/.config/notes/session` (Linux), en `0600`.

L'agent n'est donc jamais « connecté » à quoi que ce soit, et ne voit ni mot de passe ni jeton : il
appelle `create_note`, et le serveur agit déjà pour le compte. La seule chose à faire est un
`notes login`, **une fois**, sur la machine et sous l'utilisateur Unix qui lancera l'agent :

```sh
notes login
```

- La session est **par utilisateur Unix**, pas par machine : si l'agent tourne sous un utilisateur de
  service ou sous `root`, c'est sous celui-là qu'il faut se connecter.
- Elle dure **30 jours glissants**, chaque usage la relance : un agent qui s'en sert régulièrement ne
  la perd pas. Expirée, les outils répondent `not logged in, or the session expired`, et il suffit de
  refaire `notes login`.
- `notes logout` la supprime : à ne pas lancer sur une machine dont un agent dépend.

### Quand on ne peut pas se connecter là où l'agent tourne

Le jeton n'est lié à aucune machine : la base n'en garde que le hash, et `readsession` ne compare que
ça. Il se recopie donc, de la machine où la session existe :

```sh
cat ~/Library/Application\ Support/notes/session
```

vers celle de l'agent, sous le bon utilisateur :

```sh
mkdir -p ~/.config/notes && printf '%s' '<le-jeton>' > ~/.config/notes/session && chmod 600 ~/.config/notes/session
```

Les deux partagent alors la même session, et un `notes logout` d'un côté la coupe des deux.

Ce fichier ouvre le compte entier : il se traite comme un mot de passe, jamais dans un dépôt ni dans
un fichier de configuration versionné. Il n'y a pas encore de jeton par variable d'environnement, ce
qui serait plus pratique pour un conteneur ou un agent hébergé.

### Ce que ça implique

Tout agent capable de lancer `notes mcp` sur cette machine a **tous les droits du compte connecté**,
écriture et suppression comprises. La garantie est celle de l'utilisateur Unix, exactement comme pour
`notes show note`. Il n'y a ni compte réservé aux agents, ni droits en lecture seule, ni partage
entre comptes : c'est le compte, en entier.

## 3. Ce qu'il peut faire

À la connexion, l'agent reçoit les vingt-deux outils avec leur description et leur schéma, puis travaille
en langage naturel : « range mes notes de réunion de novembre dans /Travail/Réunions » devient un
`search_notes`, un `create_folder`, puis une suite de `move_item`.

| Outil | Ce qu'il fait |
|---|---|
| `search_notes` | Les notes dont le titre ou le contenu contient un texte, avec la ligne trouvée |
| `recent_notes` | Les notes modifiées le plus récemment, la plus récente d'abord |
| `read_note` | Le contenu d'une note, son chemin et ses tags |
| `list_folder` | Les notes et dossiers directement dans un dossier |
| `tags` | Les notes portant un tag, avec leur chemin ; sans tag, les tags du compte, les plus utilisés d'abord |
| `create_note` | Une note à partir de markdown : le premier titre devient le titre, les `#hashtags` des tags |
| `edit_note` | Remplace **tout** le contenu d'une note ; refuse une note verrouillée |
| `create_folder` | Un dossier, à la racine ou dans un autre |
| `rename_folder` | Le titre d'un dossier |
| `move_item` | Déplace une note ou un dossier ; un dossier emmène son contenu |
| `delete_note` | Met une note à la corbeille |
| `delete_folder` | Met un dossier **et tout son contenu** à la corbeille |
| `list_trash` | Ce qui est à la corbeille, avec l'uuid que `restore_item` attend |
| `restore_item` | Ressort de la corbeille, avec ce qui y est parti en même temps |
| `rename_tag` | Renomme un tag sur toutes ses notes, dans leur texte aussi |
| `delete_tag` | Retire un tag de toutes ses notes ; une note qui ne contenait que lui part à la corbeille |
| `note_history` | Les versions antérieures d'une note (sans `version`), ou le texte de l'une d'elles |
| `note_links` | Les liens d'une note, ses rétroliens et les notes qui partagent ses tags ; sans note, les liens cassés de tout le compte |
| `read_file` | Une image (que l'agent voit) ou une pièce jointe citée par une note ; sans fichier, la liste des fichiers |
| `attach_file` | Stocke un fichier envoyé en base64 et rend le markdown qui le cite, à poser avec `create_note` ou `edit_note` |
| `export_notes` | Un dossier, des éléments ou tout le compte en fichiers markdown, `[[liens]]` réécrits |
| `import_notes` | Des fichiers markdown en notes, leurs répertoires en dossiers, tout ou rien |

**Un outil par domaine, pas un par processus.** Chaque outil annoncé coûte du contexte à l'agent à
chaque connexion (environ 300 à 750 octets de nom, description et schéma d'entrée ; le schéma de
sortie, que les clients ne donnent en général pas au modèle, double à peu près ce chiffre). Lister et
lire passent donc par le même outil, un argument vide donnant la liste (`tags`, `note_history`,
`note_links`, `read_file`) : `list_tags` et `read_tag` sont devenus `tags`, et les treize processus
ajoutés ensuite tiennent en huit outils. La création d'un tag seul n'est pas exposée,
un tag sans note ne sert à rien à un agent.
`delete_note` et `delete_folder` restent séparés exprès : un chemin de dossier donné à `delete_note`
est refusé, au lieu d'envoyer tout un sous-arbre à la corbeille.

Ils passent par les mêmes processus (`app/shared/services/process/`) que la CLI et l'API : une règle
(titre, tags, liens, corbeille, chiffrement) s'applique pareil des trois côtés.

### Ce qu'il faut savoir avant de le lâcher dessus

- **Une note ou un dossier se désigne par son uuid ou par son chemin** depuis la racine
  (`/Dossier/Note`). Seul `restore_item` exige l'uuid : un élément à la corbeille n'a plus de chemin à
  résoudre. Un chemin qui désigne deux éléments est refusé (`several items at this path, use a uuid`).
- **`edit_note` écrase.** Rien n'est conservé de l'ancien contenu, l'agent doit lire la note avant. Le
  passage est enregistré dans l'historique de la note, qui permet de revenir en arrière depuis l'app.
- **Une suppression passe par la corbeille**, d'où `restore_item` la ressort. Mais la corbeille se
  vide d'elle-même passé le délai de rétention du compte (`purgeitems`, lancé par le serveur et à
  chaque commande de la CLI), sauf si ce compte garde sa corbeille indéfiniment : ce n'est pas un
  filet garanti. Le vidage définitif, lui, n'est pas donné à
  l'agent : il reste à la CLI et à l'app.
- **Un agent réorganise vite.** `delete_folder` envoie tout un sous-arbre à la corbeille en un appel,
  et `move_item` ne demande aucune confirmation. Un client MCP qui demande validation avant chaque
  appel d'outil est la bonne façon de travailler au début. `make backup` avant une grosse
  réorganisation reste la règle.
- **Les fichiers passent par Silo.** `read_file` (lecture), `attach_file` et `import_notes` ouvrent
  Silo à l'appel : sans les variables `S3_*`, ils répondent en erreur, les autres outils marchent.
  `read_file` rend le fichier entier en un bloc, une vidéo comprise.
- **`delete_tag` et `rename_tag` réécrivent le texte des notes**, corbeille comprise, sans passer par
  l'historique : c'est le comportement de la CLI et de l'app.
- **Ce qui n'est pas exposé** : le graphe complet (`note_links` en donne le voisinage note par note),
  les plans, la gestion des comptes, la création d'un tag seul, le vidage de la corbeille.

## Vérifier que ça marche

Le test `app/cli/cmd-mcp/cmd_test.go` branche un client MCP en mémoire sur le serveur et vérifie que
les vingt-deux outils sont annoncés avec un schéma et une description — ce que le SDK ne sait pas déduire
seul d'une structure d'arguments mal formée.

```sh
make test ARGS=./app/cli/cmd-mcp/
```

Côté client, `claude mcp list` dit si le serveur répond. S'il échoue, c'est presque toujours
l'environnement : base injoignable, `NOTES_SECRET` absent, ou pas de session (`notes login`).
