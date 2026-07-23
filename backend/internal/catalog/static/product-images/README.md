# Imagens de produto (disco + embed)

Uploads do painel admin são gravados **nesta pasta** com o nome `{slug-do-produto}.{ext}` (ex.: `macarrao-espaguete-500g.webp`).

- **API local** (`make api`): `UPLOAD_DIR=internal/catalog/static` → arquivos aqui no repositório.
- **Docker Compose**: por padrão **sem** bind-mount de `static/` na API (usa embed + `COPY` na imagem). Para uploads admin no host, monte só `product-images/` — ver `compose.yaml`.

O binário também embute os arquivos presentes no build (`//go:embed`); em runtime a API lê **primeiro o disco** (esta pasta) e depois o embed.
