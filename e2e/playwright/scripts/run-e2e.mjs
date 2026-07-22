#!/usr/bin/env node
import { spawnSync } from 'node:child_process';

if (process.env.E2E_RUN !== '1') {
  console.error('Defina E2E_RUN=1 (ex.: E2E_RUN=1 pnpm test:e2e)');
  process.exit(1);
}

function run(cmd, args) {
  const r = spawnSync(cmd, args, { stdio: 'inherit', shell: false });
  if (r.status !== 0) {
    process.exit(r.status ?? 1);
  }
}

run('pnpm', ['exec', 'playwright', 'install', 'chromium']);
run('pnpm', ['exec', 'playwright', 'test']);
