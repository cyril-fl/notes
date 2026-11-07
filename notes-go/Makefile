.PHONY: setup test backup api openapi

# Not `go tool lefthook install`: the hook would call a build cache path that expires, then skip its checks without failing.
setup:
	go install github.com/evilmartians/lefthook/v2
	"$$(go env GOBIN | grep . || echo "$$(go env GOPATH)/bin")/lefthook" install --force

# The tests run in <POSTGRES_DB>_test, made on first use: they create the schema, a migration being written must not
# reach the base the API serves. -p 1: the tests of every package share it, and deadlock when they create the schema
# and tags at the same time.
test:
	@set -a && . ./.env && set +a && \
	docker compose exec -T postgres sh -c 'psql -U "$$POSTGRES_USER" -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = '"'"'$${POSTGRES_DB}_test'"'"'" | grep -q 1 || createdb -U "$$POSTGRES_USER" "$${POSTGRES_DB}_test"' && \
	POSTGRES_DB=$${POSTGRES_DB}_test CONFIG_FILE=$(CURDIR)/config/config.yaml go test -p 1 $(or $(ARGS),./...)

# Restore: docker compose exec -T postgres sh -c 'psql -U "$$POSTGRES_USER" "$$POSTGRES_DB"' < backups/<file>.sql
backup:
	@mkdir -p backups
	@file=backups/notes-$$(date +%Y%m%d-%H%M%S).sql; \
	docker compose exec -T postgres sh -c 'pg_dump --clean --if-exists -U "$$POSTGRES_USER" "$$POSTGRES_DB"' > $$file.tmp \
		&& mv $$file.tmp $$file && echo $$file \
		|| { rm -f $$file.tmp; exit 1; }

api:
	@docker compose up -d --build api

# Generates docs/swagger.yaml from the @ comments of app/api; no docs.go, the binary stays out of it.
# Every field is required unless tagged binding:"optional".
openapi:
	go tool swag init --v3.1 --requiredByDefault -d app/api,app/shared -g router.go -o docs --outputTypes yaml

up:
	@colima start && docker compose up -d

down:
	@docker compose down && colima stop