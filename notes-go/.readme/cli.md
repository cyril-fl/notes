# Spécification de la CLI — Notes

## Périmètre

Ce document définit l'interface en ligne de commande du projet :

* Noms et hiérarchie des commandes.
* Arguments positionnels.
* Options et flags.
* Comportements par défaut du point de vue de l'utilisateur.
* Formats de sortie.
* Conventions d'utilisation.
* Historique des décisions relatives à l'interface.

La logique métier, les règles de calcul des relations, le modèle de stockage et les détails d'implémentation ne sont pas définis ici.

## Décisions de nomenclature

| Opération              | Commande        |
| ---------------------- | --------------- |
| Créer une note         | `note add`      |
| Afficher une note      | `note show`     |
| Modifier une note      | `note update`   |
| Déplacer une note      | `note move`     |
| Supprimer une note     | `note remove`   |
| Créer un dossier       | `folder add`    |
| Afficher un dossier    | `folder show`   |
| Modifier un dossier    | `folder update` |
| Déplacer un dossier    | `folder move`   |
| Supprimer un dossier   | `folder remove` |
| Créer un tag           | `tag add`       |
| Afficher un tag        | `tag show`      |
| Modifier un tag        | `tag update`    |
| Supprimer un tag       | `tag remove`    |
| Rechercher et naviguer | `finder`        |
| Explorer les relations | `graph`         |

## Conventions retenues

* Le label d'une note correspond à son titre, indépendamment de son chemin.
* Le label d'un dossier correspond à son nom.
* Les commandes peuvent accepter un identifiant ou un chemin lorsque cela est pertinent.
* Le Finder est interactif par défaut, avec une option de sortie en liste.
* La navigation interactive dans les dossiers n'est pas prévue pour la première version.
* Les sorties sont lisibles par défaut et peuvent être fournies en JSON avec `--json`.
* Les suppressions demandent une confirmation, sauf avec `--force`.
* `related` désigne l'affichage des notes liées ; `links` n'est pas retenu.
* `--min-weight` permet de définir un seuil minimal pour les résultats de `related` et de `graph`, avec une valeur par défaut de `1`.
* `graph --folder` permet de limiter le périmètre à un dossier, avec une portée récursive par défaut et une option `--direct`.

## Structure des commandes

```text
scn
├── note
│   ├── add [title]
│   ├── show <id-or-path>
│   ├── update <id>
│   ├── move <id> <destination>
│   └── remove <id-or-path>
├── folder
│   ├── add <name>
│   ├── show <id-or-path>
│   ├── update <id>
│   ├── move <id> <destination>
│   └── remove <id-or-path>
├── tag
│   ├── add <value>
│   ├── show <id-or-value>
│   ├── update <id> --value <value>
│   └── remove <id>
├── finder [query]
└── graph
```

## Options principales

### Notes

* `note add`: `--description`, `--path` et association de tags.
* `note show`: `--related`, `--min-weight`, `--json`.
* `note update`: `--label`, `--description`, `--add-tag`, `--remove-tag`.
* `note move`: destination explicite.
* `note remove`: `--force`.

### Dossiers

* `folder add`: `--parent`, `--description`.
* `folder show`: `--tree`, `--recursive`, `--json`.
* `folder update`: `--label`, `--description`.
* `folder move`: destination explicite.
* `folder remove`: `--force`.

### Tags

* `tag add`: valeur du tag.
* `tag show`: identifiant ou valeur.
* `tag update`: `--value`.
* `tag remove`: `--force`.

### Finder

* `-l`, `--list`.
* `-i`, `--interactive`.
* `-t`, `--tag`.
* `-p`, `--path`.
* `--type`.
* `--json`.

### Graph

* `-f`, `--folder`.
* `--min-weight`.
* `--direct`.
* `--json`.

## Hors périmètre

Ce document ne définit pas :

* La formule de calcul du poids des relations.
* Le fonctionnement interne du graphe.
* Les algorithmes de recherche.
* Le modèle de données et le stockage.
* L'architecture interne de l'application.
* Les règles internes de calcul des statistiques.
* L'implémentation des interfaces interactives.

Ces sujets relèvent de la conception métier ou technique, et doivent être documentés séparément si nécessaire.
