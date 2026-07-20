# Portainer + Traefik (VPS compartilhada)

Use o Portainer para **monitorar** a stack; o **deploy de versão** deve vir do CI (`deploy.sh` via SSH) para evitar drift entre UI e Git.

## Pré-requisitos

1. Traefik já rodando na VPS (Docker provider ativo).
2. Rede Docker externa do Traefik (ex.: `traefik`) — veja em *Containers → traefik → Networks*.
3. Anotar no `.env.production`:
   - `TRAEFIK_NETWORK`
   - `TRAEFIK_ENTRYPOINT` (ex.: `websecure`)
   - `TRAEFIK_CERT_RESOLVER` (ex.: `letsencrypt`)

## Criar a stack

1. Portainer → **Stacks** → **Add stack** → nome `store-platform`.
2. **Web editor**: cole o conteúdo de [`../compose/compose.traefik.yaml`](../compose/compose.traefik.yaml) **ou** use *Git repository* apontando para este repo (caminho do compose: `infra/compose/compose.traefik.yaml`).
3. **Environment variables**: importe de `infra/compose/.env.production` (não commitar secrets).
4. **Deploy the stack**.

Na primeira vez, prefira deploy pelo servidor:

```bash
cd /opt/store-platform
export COMPOSE_FILE=infra/compose/compose.traefik.yaml
./infra/deploy/deploy.sh
```

Assim o `git pull` e o build ficam alinhados ao pipeline.

## Rede Traefik

A stack **não** publica portas 80/443. Os serviços `store-web`, `admin-web` e `api` (rota `/health`) entram na rede externa `${TRAEFIK_NETWORK}`.

Se o Traefik não enxergar os containers:

- Confirme `traefik.docker.network` nas labels (já definido no compose).
- Confirme que o Traefik está na mesma rede Docker.

## Rotas esperadas

| Host | Destino |
| ---- | ------- |
| `DOMAIN` | `store-web:80` (nginx faz proxy `/api` → `api`) |
| `ADMIN_DOMAIN` | `admin-web:80` |
| `DOMAIN` + `Path(/health)` | `api:8080` (prioridade alta) |

## VPS dedicada (sem Traefik)

Use [`../compose/compose.production.yaml`](../compose/compose.production.yaml) com Caddy e `COMPOSE_FILE` apontando para esse arquivo.

## Portainer CE standalone (opcional)

Se ainda não tiver Portainer:

```bash
docker volume create portainer_data
docker run -d -p 9443:9443 --name portainer \
  --restart=always \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v portainer_data:/data \
  portainer/portainer-ce:latest
```
