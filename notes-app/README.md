# notes-app

Web app de `notes-go`, en Nuxt et Nuxt UI, sans serveur : le navigateur appelle directement l'API.

## Lancer

Toute la stack, depuis la racine du repo : `make up`, puis `make down` pour l'arrêter (Colima reste démarré, d'autres projets peuvent l'utiliser). `make up` lance la web app en arrière-plan, avec ses logs dans `.build/notes-app.log`.

Ou pièce par pièce :

Dans `notes-go/.env`, autorise l'app :

```sh
WEB_ORIGIN=http://localhost:3000
```

```sh
cd ../notes-go && make up && make api   # PostgreSQL et l'API sur :8080
cd ../notes-app && pnpm install && pnpm dev   # http://localhost:3000
```

Ouvre bien `localhost`, pas `127.0.0.1` : l'API compare l'origine exactement. `NUXT_PUBLIC_API_BASE` change l'adresse de l'API.

## Raccourcis

`⌘K` ou `/` ouvre la palette (notes, récentes, actions, thème), `N` crée une note dans le dossier affiché, `⇧N` un dossier, `⌘\` replie la barre latérale. Hors d'un champ de saisie seulement.

La barre latérale se redimensionne en tirant son bord (sa largeur reste d'une visite à l'autre). Repliée, elle reste à l'écran en barre d'icônes (recherche, notes, graphe, corbeille, réglages, thème). Sur petit écran, elle devient un tiroir.

Un clic droit sur un dossier le met en favori, un clic droit sur un tag le renomme ou le supprime.

Dans un dossier, ⌘-clic (Ctrl hors macOS) sélectionne plusieurs éléments et ⇧-clic une plage : une barre propose de les déplacer ou de les supprimer, et glisser l'un d'eux emmène toute la sélection. Échap annule la sélection.

Le graphe se déplace en glissant et zoome à la molette, en pinçant le trackpad ou avec ses boutons + et −. « Recentrer » le remet comme au départ.

## Réglages

La page Réglages (roue dentée en bas de la barre latérale, ou la palette) regroupe ce qui touche à l'app entière. Pour l'instant :

- **Exporter** télécharge toutes les notes en zip de fichiers `.md`, comme `notes export`.
- **Importer** ajoute les `.md` d'un zip au dossier choisi (la racine par défaut), comme `notes import --parent`. Si un dossier du zip existe déjà à cet endroit, rien n'est importé.

## Commandes

```sh
pnpm test        # file de sauvegarde, liens et disposition du graphe
pnpm e2e         # parcours dans Chrome, contre la stack lancée (make up) ; ou make e2e depuis la racine
pnpm typecheck
pnpm types       # régénère app/types/api.d.ts depuis ../notes-go/docs/swagger.yaml
pnpm generate    # site statique dans .output/public
```
