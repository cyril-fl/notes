.PHONY: setup test

setup:
	go tool lefthook install

test:
	set -a && . ./.env && set +a && CONFIG_FILE=$(CURDIR)/config/config.yaml go test $(or $(ARGS),./...)
