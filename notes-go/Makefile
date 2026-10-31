.PHONY: setup test

setup:
	go tool lefthook install

test:
	set -a && . ./.env && set +a && CONFIG_FILE=$(CURDIR)/config/config.yaml go test $(or $(ARGS),./...)

up:
	@colima start && docker compose up -d

down:
	@docker compose down && colima stop