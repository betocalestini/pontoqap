## Descrição

<!-- O que mudou e por quê -->

## Checklist — portão automatizado (fase 0)

Marque após rodar localmente ou confirmar CI verde. Detalhes em [`docs/homologation.md`](../docs/homologation.md) e [`docs/development/testing.md`](../docs/development/testing.md).

- [ ] `make migrate-up` (se schema mudou) e `make test`
- [ ] `cd backend && go test -p 1 ./tests/e2e/... -run 'TestHTTP|TestMVP'` (ou job `backend-e2e` no CI)
- [ ] `pnpm test` (quando alterou frontend ou packages TS)
- [ ] `pnpm -r run build` (apps afetados)
- [ ] `make test-backup-restore` (se alterou migrations/backup; senão confiar no nightly)

## Test plan

<!-- Passos manuais se aplicável -->
