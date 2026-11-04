# notes-app

Web app de `notes-go`, en Nuxt et Nuxt UI, sans serveur : le navigateur appelle directement l'API.

## Lancer

Dans `notes-go/.env`, autorise l'app :

```sh
WEB_ORIGIN=http://localhost:3000
```

```sh
cd ../notes-go && make up && make api   # PostgreSQL et l'API sur :8080
cd ../notes-app && pnpm install && pnpm dev   # http://localhost:3000
```

Ouvre bien `localhost`, pas `127.0.0.1` : l'API compare l'origine exactement. `NUXT_PUBLIC_API_BASE` change l'adresse de l'API.

## Commandes

```sh
pnpm test        # file de sauvegarde
pnpm typecheck
pnpm types       # régénère app/types/api.d.ts depuis ../notes-go/docs/swagger.yaml
pnpm generate    # site statique dans .output/public
```
