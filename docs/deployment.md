# Implantação

Referência para staging/produção (decisions.md §19).

## Stack de produção

1. Copie `infra/compose/.env.production.example` para o host e preencha secrets.
2. Ajuste `infra/caddy/Caddyfile` (`DOMAIN`, `ACME_EMAIL`).
3. Suba a stack:

```bash
docker compose -f infra/compose/compose.production.yaml --env-file infra/compose/.env.production up -d --build
```

## Checklist

- [ ] `ADMIN_MFA_REQUIRED=true` e gerente com MFA ativo
- [ ] `COOKIE_SECURE=true` e TLS via Caddy
- [ ] Backup diário (`cron` + `make backup`)
- [ ] `infra/backup/verify_restore.sh` executado após mudanças de schema
- [ ] Homologação (`docs/homologation.md`)

## CI/CD

- `pull-request.yml` — testes em PR
- `build-images.yml` — imagens Docker no registry
- `deploy-staging.yml` — deploy manual (workflow_dispatch) com secrets do ambiente

## Portainer

Ver `infra/portainer/README.md` para importar o compose no servidor.
