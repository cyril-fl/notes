.PHONY: up down e2e perf

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

# The app at scale (#138): the account perf of 50,000 generated notes, made once (14 s), then the timings of its API
# and pages, a table to paste in an issue. PERF_NOTES only counts when the account does not exist yet.
PERF_PASSWORD := perf-password
PERF_NOTES := 50000
perf: up
	@echo "$(PERF_PASSWORD)" | docker exec -i note-api /notes user perf perf $(PERF_NOTES) --password-stdin
	@cd notes-app && PERF_PASSWORD=$(PERF_PASSWORD) node scripts/perf.mjs
