.PHONY: setup test backup

setup:
	go tool lefthook install

test:
	set -a && . ./.env && set +a && CONFIG_FILE=$(CURDIR)/config/config.yaml go test $(or $(ARGS),./...)

# Restore: docker compose exec -T postgres sh -c 'psql -U "$$POSTGRES_USER" "$$POSTGRES_DB"' < backups/<file>.sql
backup:
	@mkdir -p backups
	@file=backups/notes-$$(date +%Y%m%d-%H%M%S).sql; \
	docker compose exec -T postgres sh -c 'pg_dump --clean --if-exists -U "$$POSTGRES_USER" "$$POSTGRES_DB"' > $$file.tmp \
		&& mv $$file.tmp $$file && echo $$file \
		|| { rm -f $$file.tmp; exit 1; }

up:
	@colima start && docker compose up -d

down:
	@docker compose down && colima stop