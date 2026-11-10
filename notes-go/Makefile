.PHONY: setup test backup purge api openapi pgsql

# Not `go tool lefthook install`: the hook would call a build cache path that expires, then skip its checks without failing.
setup:
	go install github.com/evilmartians/lefthook/v2
	go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
	"$$(go env GOBIN | grep . || echo "$$(go env GOPATH)/bin")/lefthook" install --force

# The tests run in <POSTGRES_DB>_test, made on first use: they create the schema, a migration being written must not
# reach the base the API serves. -p 1: the tests of every package share it, and deadlock when they create the schema
# and tags at the same time.
test:
	@set -a && . ./.env && set +a && \
	docker compose exec -T postgres sh -c 'psql -U "$$POSTGRES_USER" -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = '"'"'$${POSTGRES_DB}_test'"'"'" | grep -q 1 || createdb -U "$$POSTGRES_USER" "$${POSTGRES_DB}_test"' && \
	POSTGRES_DB=$${POSTGRES_DB}_test CONFIG_FILE=$(CURDIR)/config/config.yaml go test -p 1 $(or $(ARGS),./...)

# The base, then the images of Silo beside it. Restore:
#   docker compose exec -T postgres sh -c 'psql -U "$$POSTGRES_USER" "$$POSTGRES_DB"' < backups/<file>.sql
#   docker compose cp backups/<file>.files silo:/tmp/restore && docker compose exec silo sh -c \
#     'mcli alias set local http://127.0.0.1:9000 "$$MINIO_ROOT_USER" "$$MINIO_ROOT_PASSWORD" && mcli mirror /tmp/restore local/<bucket>'
backup:
	@mkdir -p backups
	@set -a && . ./.env && set +a && \
	file=backups/notes-$$(date +%Y%m%d-%H%M%S); \
	docker compose exec -T postgres sh -c 'pg_dump --clean --if-exists -U "$$POSTGRES_USER" "$$POSTGRES_DB"' > $$file.tmp \
		&& mv $$file.tmp $$file.sql && echo $$file.sql \
		|| { rm -f $$file.tmp; exit 1; }; \
	docker compose exec -T silo sh -c 'rm -rf /tmp/backup && mkdir /tmp/backup \
		&& mcli alias set local http://127.0.0.1:9000 "$$MINIO_ROOT_USER" "$$MINIO_ROOT_PASSWORD" >/dev/null \
		&& mcli mirror --quiet local/'"$$S3_BUCKET"' /tmp/backup >/dev/null' \
		&& docker compose cp silo:/tmp/backup $$file.files && echo $$file.files

# Every note, folder, tag and image of every account, after a backup. The accounts, their settings and sessions stay.
# One account empties its own from the settings of the web app (DELETE /api/me/data).
purge:
	@printf 'Delete the notes, folders, tags and images of every account? Type purge: ' && read answer && [ "$$answer" = purge ]
	@$(MAKE) --no-print-directory backup
	@set -a && . ./.env && set +a && \
	docker compose exec -T postgres sh -c 'psql -v ON_ERROR_STOP=1 -U "$$POSTGRES_USER" "$$POSTGRES_DB" -c "TRUNCATE items, tags, files CASCADE" -c "UPDATE users SET avatar = NULL"' && \
	docker compose exec -T silo sh -c 'mcli alias set local http://127.0.0.1:9000 "$$MINIO_ROOT_USER" "$$MINIO_ROOT_PASSWORD" >/dev/null \
		&& mcli rm --recursive --force --quiet local/'"$$S3_BUCKET"

# psql on the base: `make pgsql` opens it, `make pgsql QUERY="ALTER TABLE ..."` runs one statement and stops at an error.
pgsql:
	@set -a && . ./.env && set +a && \
	if [ -n "$$QUERY" ]; then \
		docker compose exec -T -e QUERY postgres sh -c 'psql -v ON_ERROR_STOP=1 -U "$$POSTGRES_USER" "$$POSTGRES_DB" -c "$$QUERY"'; \
	else \
		docker compose exec postgres sh -c 'psql -U "$$POSTGRES_USER" "$$POSTGRES_DB"'; \
	fi

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