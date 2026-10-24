.PHONY: up down run stop

RUN_DIR := /tmp/doratou-$(shell id -u)-$(shell pwd | cksum | cut -d' ' -f1)

up:
	docker compose up -d --wait

down:
	docker compose down

run:
	@mkdir -p "$(RUN_DIR)"
	@if [ -f "$(RUN_DIR)/app.pid" ] && kill -0 "$$(cat "$(RUN_DIR)/app.pid")" 2>/dev/null; then echo "Application is already running"; exit 1; fi
	go build -o "$(RUN_DIR)/app" ./cmd/app
	@echo $$$$ > "$(RUN_DIR)/app.pid"; exec "$(RUN_DIR)/app"

stop:
	@if [ -f "$(RUN_DIR)/app.pid" ]; then \
		pid=$$(cat "$(RUN_DIR)/app.pid"); \
		if [ "$$(readlink "/proc/$$pid/exe")" = "$$(realpath "$(RUN_DIR)/app")" ]; then kill "$$pid"; fi; \
		rm -f "$(RUN_DIR)/app.pid"; \
	fi
