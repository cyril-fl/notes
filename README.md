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