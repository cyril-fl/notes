.PHONY: up down e2e

WEB_PORT := 3000
WEB_LOG := $(CURDIR)/.build/notes-app.log

# Docker, API, then the web app in the background.
up:
	@$(MAKE) -C notes-go up api
	@mkdir -p .build
	@if lsof -ti tcp:$(WEB_PORT) -sTCP:LISTEN >/dev/null; then \
		echo "Port $(WEB_PORT) already in use: web app not started"; \
	else \
		cd notes-app && bun install --silent && \
		{ nohup bun run dev > $(WEB_LOG) 2>&1 < /dev/null & } && \
		echo "http://localhost:$(WEB_PORT) (log: .build/notes-app.log)"; \
	fi

# Keeps Colima running: other projects may use it. --profile api: a plain down leaves the API out.
# TODO: stops whatever listens on the web port, add a pid file if another app ever uses it.
down:
	@lsof -ti tcp:$(WEB_PORT) -sTCP:LISTEN | xargs kill 2>/dev/null || true
	@cd notes-go && docker compose --profile api down

# The browser tests run against the stack, in the installed Chrome.
e2e: up
	@cd notes-app && bun run e2e
