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

## Qualité

- `make setup` dans `notes-go/` installe les hooks git (lefthook) : gofmt et Prettier avant un commit, et le préfixe du message (`feat:`, `fix:`…) vérifié.
- La CI GitHub Actions lance à chaque push les tests Go, le lint, le format, les tests, le typecheck et le build de l'app.
