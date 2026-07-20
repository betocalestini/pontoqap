GO ?= go
PNPM ?= pnpm

.PHONY: help deps migrate-up api worker test lint docker-up docker-down

help:
	@echo "Targets: deps, migrate-up, api, worker, test, docker-up"

deps:
	cd backend && $(GO) mod download
	$(PNPM) install

migrate-up:
	cd backend && $(GO) run ./cmd/migrate -direction up

api:
	cd backend && $(GO) run ./cmd/api

worker:
	cd backend && $(GO) run ./cmd/worker

test:
	cd backend && $(GO) test ./...

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down
