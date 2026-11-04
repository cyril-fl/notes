.PHONY: setup test backup api openapi

# Not `go tool lefthook install`: the hook would call a build cache path that expires, then skip its checks without failing.
setup:
	go install github.com/evilmartians/lefthook/v2
	"$$(go env GOBIN | grep . || echo "$$(go env GOPATH)/bin")/lefthook" install --force

# -p 1: the tests of every package share one database, and deadlock when they create the schema and tags at the same time.
test:
	set -a && . ./.env && set +a && CONFIG_FILE=$(CURDIR)/config/config.yaml go test -p 1 $(or $(ARGS),./...)

# Restore: docker compose exec -T postgres sh -c 'psql -U "$$POSTGRES_USER" "$$POSTGRES_DB"' < backups/<file>.sql
backup:
	@mkdir -p backups
	@file=backups/notes-$$(date +%Y%m%d-%H%M%S).sql; \
	docker compose exec -T postgres sh -c 'pg_dump --clean --if-exists -U "$$POSTGRES_USER" "$$POSTGRES_DB"' > $$file.tmp \
		&& mv $$file.tmp $$file && echo $$file \
		|| { rm -f $$file.tmp; exit 1; }

# Not vendor/: Go would then use it for every local build, and ignore changes to plum.
api:
	@go mod vendor -o .docker-vendor
	@docker compose up -d --build api; status=$$?; rm -rf .docker-vendor; exit $$status

# Generates docs/swagger.yaml from the @ comments of app/api; no docs.go, the binary stays out of it.
openapi:
	go tool swag init --v3.1 -d app/api,app/shared -g router.go -o docs --outputTypes yaml

up:
	@colima start && docker compose up -d

down:
	@docker compose down && colima stop