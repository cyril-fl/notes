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
cd ../notes-go && make up && make api   # PostgreSQL, Silo et l'API sur :8080
cd ../notes-app && bun install && bun run dev   # http://localhost:3000
```

Ouvre bien `localhost`, pas `127.0.0.1` : l'API compare l'origine exactement. `NUXT_PUBLIC_API_BASE` change l'adresse de l'API.

L'app demande de se connecter avec un compte créé par la CLI (`notes user add <nom>`, voir le README de `notes-go`). Sans session, elle ouvre la page de connexion, puis revient à la page demandée. L'avatar en bas de la barre latérale, réduite ou non, ouvre un menu : Compte, Réglages, Se déconnecter. La page Compte change l'avatar (une image du compte, les initiales sans elle), le nom et le mot de passe. Changer le mot de passe déconnecte les autres appareils.

## Images

Coller ou glisser une image (PNG, JPEG, GIF, WebP, 20 Mo au plus) dans une note l'envoie à l'API et l'insère à cet endroit. Elle est rangée dans Silo, voir le README de `notes-go`.

## Raccourcis

`⌘K` ou `/` ouvre la palette (notes, récentes, actions, thème), `N` crée une note dans le dossier affiché, `T` une note qui commence par un snippet, `⇧N` un dossier, `⌘\` replie la barre latérale. Hors d'un champ de saisie seulement.

La barre latérale se redimensionne en tirant son bord (sa largeur reste d'une visite à l'autre). Repliée, elle reste à l'écran en barre d'icônes (recherche, notes, graphe, corbeille, réglages, et le bouton clair/sombre sauf en mode système, où l'OS décide). Sur petit écran, elle devient un tiroir.

Un clic droit sur un dossier le met en favori, un clic droit sur un tag le renomme ou le supprime.

Dans un dossier, ⌘-clic (Ctrl hors macOS) sélectionne plusieurs éléments et ⇧-clic une plage : une barre propose de les déplacer ou de les supprimer, et glisser l'un d'eux emmène toute la sélection. Échap annule la sélection.

Le graphe se déplace en glissant et zoome à la molette, en pinçant le trackpad ou avec ses boutons + et −. « Recentrer » le remet comme au départ.

## Réglages

La page Réglages (menu de l'avatar en bas de la barre latérale, ou la palette) regroupe ce qui touche à l'app entière. Les réglages sont rangés sur le compte (`PATCH /api/me`) : ils suivent d'un appareil à l'autre. Ils sont rangés par sections (Général, Apparence, Éditeur, Données, Mode Pro), chacune avec sa zone de danger en bas quand elle en a une (Données : tout effacer), listées dans `app/utils/settings.ts`. En mode Pro, la barre latérale devient la navigation des réglages (Compte puis une page par section, `/settings/<section>`), avec un lien de retour aux notes.

- **Langue** de l'interface, français ou anglais.
- **Apparence** : le thème, choisi d'un clic sur sa miniature parmi des jeux de couleurs tout faits (Notes, Dracula, Catppuccin, Nord, Gruvbox, Everforest, Tokyo Night, Vue), et le mode (système, clair, sombre). Un thème se déclare dans `app/utils/themes.ts`, ses palettes dans `main.css`, et son id dans `domain.Themes` côté Go. Le bouton de mode de la barre latérale enregistre aussi, comme celui de la vue des listes.
- **Mode Pro** : des fonctionnalités en plus, listées sous l'interrupteur ; le mode focus (`⌘.` ou le bouton en bas d'une note), qui masque la barre latérale jusqu'à changer de page, et sur grand écran un panneau à droite de la note avec son plan (un clic sur un titre y fait défiler), le graphe de ses voisins et ses notes liées, et le nombre de mots en bas de la note, mis à jour à la frappe. Le clic droit (ou l'appui long) dans une note ouvre les commandes du menu `/`, les encadrés dans un sous-menu. Le niveau de chaque fonctionnalité (`base` pour tous, `pro`, `off` pour personne) se déclare dans `app/utils/features.ts` : l'app les montre avec `usePro().has(…)`, et une page déclarée `definePageMeta({ feature })` renvoie à l'accueil quand elle n'est pas disponible. Le mode Pro ne touche que l'interface : l'API répond pareil.
- **Snippets** : du markdown nommé, dont les `#tags` deviennent les tags de la note où il arrive. Il s'insère au curseur en tapant `;;` dans une note, puis son nom, ou donne le texte de départ d'une nouvelle note (`T`, ou « Nouvelle note depuis un snippet » dans la palette) ; cette note n'est créée qu'une fois qu'on y tape.
- **Corbeille** : la durée avant que les éléments supprimés soient effacés pour de bon (7, 30, 90, 365 jours, ou jamais).
- **Exporter** télécharge toutes les notes en zip de fichiers `.md`, leurs images dans `_files/`, comme `notes export`.
- **Importer** ajoute les `.md` d'un zip au dossier choisi (la racine par défaut), comme `notes import --parent`. Si un dossier du zip existe déjà à cet endroit, rien n'est importé.

## Commandes

```sh
bun run test        # file de sauvegarde, liens et disposition du graphe
bun run e2e         # parcours dans Chrome, contre la stack lancée (make up) ; ou make e2e depuis la racine
                    # se connecte avec le compte e2e (E2E_NAME, E2E_PASSWORD) : notes user add e2e, après le tien
                    # E2E_APP et E2E_API pointent les tests vers une autre stack
bun run typecheck
bun run types       # régénère app/types/api.d.ts depuis ../notes-go/docs/swagger.yaml
bun run generate    # site statique dans .output/public
```
