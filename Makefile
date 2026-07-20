GO ?= go
PNPM ?= pnpm

.PHONY: help deps migrate-up api worker test lint docker-up docker-down openapi-gen backup restore

help:
	@echo "Targets: deps, migrate-up, api, worker, test, docker-up, openapi-gen, backup, restore"

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
	cd backend && $(GO) test -p 1 ./...

test-integration:
	cd backend && DATABASE_URL=$${DATABASE_URL:-postgres://store:store@localhost:5432/store?sslmode=disable} $(GO) test -p 1 ./tests/...

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

openapi-gen:
	$(PNPM) --filter @store/contracts generate

backup:
	@chmod +x infra/backup/backup.sh
	DATABASE_URL=$${DATABASE_URL:-postgres://store:store@localhost:5432/store?sslmode=disable} ./infra/backup/backup.sh

restore:
	@chmod +x infra/backup/restore.sh
	@test -n "$(BACKUP)" || (echo "Use: make restore BACKUP=backups/arquivo.sql.gz" && exit 1)
	DATABASE_URL=$${DATABASE_URL:-postgres://store:store@localhost:5432/store?sslmode=disable} ./infra/backup/restore.sh "$(BACKUP)"

test-backup-restore:
	@chmod +x infra/backup/verify_restore.sh
	DATABASE_URL=$${DATABASE_URL:-postgres://store:store@localhost:5432/store?sslmode=disable} ./infra/backup/verify_restore.sh
