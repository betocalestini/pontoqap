# Imagens de produto (disco + embed)

Uploads do painel admin são gravados **nesta pasta** com o nome `{slug-do-produto}.{ext}` (ex.: `macarrao-espaguete-500g.webp`).

- **API local** (`make api`): `UPLOAD_DIR=internal/catalog/static` → arquivos aqui no repositório.
- **Docker Compose**: volume `./backend/internal/catalog/static` → `/data/catalog-static` no container (mesma árvore `product-images/`).

O binário também embute os arquivos presentes no build (`//go:embed`); em runtime a API lê **primeiro o disco** (esta pasta) e depois o embed.
