# Implantação e pipeline de deploy

Guia para preparar o servidor, configurar secrets no GitHub e executar deploys de staging e produção.

## Visão geral

### VPS com Portainer + Traefik (recomendado)

```text
GitHub Actions (workflow_dispatch)
        │
        ▼ SSH + deploy.sh (COMPOSE_FILE=compose.traefik.yaml)
Servidor (/opt/store-platform)
        │
        ▼ docker compose — sem portas 80/443 na stack
postgres → migrate → api + worker + store-web + admin-web
        │
        ▼ labels Traefik (rede externa)
Traefik existente na VPS → TLS + roteamento por Host
```

### VPS dedicada (sem Traefik)

Use [`infra/compose/compose.production.yaml`](../infra/compose/compose.production.yaml) com **Caddy** nas portas 80/443.

---

## 1. Preparar o servidor (uma vez)

### 1.1 VPS já com Portainer + Traefik

- **Não** execute `install-server.sh` se Docker e Traefik já estiverem configurados.
- Clone o repositório e configure `.env` (abaixo).
- Confirme o nome da **rede Docker** do Traefik (`TRAEFIK_NETWORK`).

### 1.2 VPS nova (sem Docker)

- Ubuntu 22.04+ ou Debian 12+
- DNS: `A`/`AAAA` para `DOMAIN` e `ADMIN_DOMAIN`
- Portas 80/443 no host (Traefik ou Caddy)

```bash
sudo bash infra/deploy/install-server.sh
```

### 1.3 Repositório e variáveis

```bash
git clone https://github.com/SEU_ORG/store-platform.git /opt/store-platform
cd /opt/store-platform
cp infra/compose/.env.production.example infra/compose/.env.production
chmod +x infra/deploy/deploy.sh infra/backup/*.sh
```

Edite `infra/compose/.env.production`:

| Variável | Uso |
| -------- | --- |
| `POSTGRES_*`, `DATABASE_URL` | Banco |
| `DOMAIN`, `ADMIN_DOMAIN` | Hosts Traefik / DNS |
| `CORS_ALLOWED_ORIGINS` | Ambos os URLs HTTPS |
| `TRAEFIK_NETWORK`, `TRAEFIK_ENTRYPOINT`, `TRAEFIK_CERT_RESOLVER` | Compose Traefik |
| `SESSION_SECRET`, `CSRF_SECRET`, `ENCRYPTION_KEY` | Segurança |
| `STORE_*_IMAGE` + `USE_REGISTRY_IMAGES=true` | Opcional GHCR |

### 1.4 Primeiro deploy (Traefik)

```bash
cd /opt/store-platform
export COMPOSE_FILE=infra/compose/compose.traefik.yaml
./infra/deploy/deploy.sh
```

Validar:

- `https://SEU_DOMAIN/health` → `{"status":"ok"}`
- `https://SEU_DOMAIN` — loja
- `https://SEU_ADMIN_DOMAIN` — painel

Detalhes Portainer: [`infra/portainer/README.md`](../infra/portainer/README.md).

### 1.5 Backup agendado

```cron
0 3 * * * /opt/store-platform/infra/backup/cron-backup.sh >> /var/log/store-backup.log 2>&1
```

Teste: `make test-backup-restore`.

---

## 2. Secrets no GitHub

### Ambiente `staging`

| Secret | Descrição |
| ------ | --------- |
| `STAGING_HOST`, `STAGING_USER`, `STAGING_SSH_KEY` | SSH |
| `STAGING_REPO_DIR` | default `/opt/store-platform` |
| `STAGING_COMPOSE_FILE` | opcional; default `infra/compose/compose.traefik.yaml` |
| `STAGING_HEALTH_URL` | ex. `https://loja.staging.example.com/health` |

### Ambiente `production`

| Secret | Descrição |
| ------ | --------- |
| `PRODUCTION_*` | Idem staging |
| `PRODUCTION_COMPOSE_FILE` | opcional |
| `PRODUCTION_HEALTH_URL` | obrigatório no workflow de produção |

---

## 3. Pipelines

| Workflow | Função |
| -------- | ------ |
| `pull-request.yml` | Testes |
| `build-images.yml` | Build local; **Publish no GHCR** via `workflow_dispatch` + `push_registry=true` |
| `deploy-staging.yml` | SSH + `deploy.sh` |
| `deploy-production.yml` | SSH + `deploy.sh` + smoke |

### Imagens no GHCR (opcional)

1. Actions → **Build Images** → Run workflow → marcar **push_registry**.
2. No servidor, em `.env.production`:

```bash
USE_REGISTRY_IMAGES=true
STORE_API_IMAGE=ghcr.io/seu-org/store-platform-api:latest
STORE_WORKER_IMAGE=ghcr.io/seu-org/store-platform-worker:latest
# ... demais STORE_*_IMAGE
```

3. Próximo `deploy.sh` fará `docker compose pull` antes do build.

### Deploy staging

1. Actions → **Deploy Staging** → `ref`: `master`
2. Logs: `docker compose -f infra/compose/compose.traefik.yaml --env-file infra/compose/.env.production logs api`

---

## 4. Rollback

```bash
cd /opt/store-platform
git checkout <tag-anterior>
export COMPOSE_FILE=infra/compose/compose.traefik.yaml
./infra/deploy/deploy.sh
```

---

## 5. Checklist pós-deploy

- [ ] `/health` via Traefik
- [ ] MFA do gerente em `https://ADMIN_DOMAIN/mfa`
- [ ] Homologação (`docs/homologation.md`)
- [ ] Backup diário
- [ ] Webhook Pix: `https://DOMAIN/api/v1/webhooks/pix`

---

## 6. Desenvolvimento local

```bash
cp .env.example .env
docker compose up -d
make migrate-up
make api && make worker
pnpm dev:store && pnpm dev:admin
```
