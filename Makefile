GO ?= go
PNPM ?= pnpm

.PHONY: help deps migrate-up api worker test lint docker-up docker-down dev-up dev-up-local openapi-gen backup restore test-backup-restore deploy-staging-help

help:
	@echo "Targets: deps, migrate-up, api, worker, test, docker-up, dev-up, dev-up-local, openapi-gen, backup, restore, test-backup-restore"

dev-up:
	@chmod +x scripts/dev-up.sh
	./scripts/dev-up.sh

dev-up-local:
	@chmod +x scripts/dev-up.sh
	./scripts/dev-up.sh --local

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
	PGPASSWORD=$${PGPASSWORD:-store} DATABASE_URL=$${DATABASE_URL:-postgres://store:store@localhost:5432/store?sslmode=disable} ./infra/backup/verify_restore.sh

deploy-staging-help:
	@echo "Deploy: GitHub Actions → Deploy Staging (compose.traefik.yaml por padrão)"
	@echo "Servidor: COMPOSE_FILE=infra/compose/compose.traefik.yaml ./infra/deploy/deploy.sh"
	@echo "Ver docs/deployment.md e infra/portainer/README.md"
