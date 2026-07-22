import { test, expect } from '@playwright/test';

const runE2E = process.env.E2E_RUN === '1';

test.describe('smoke', () => {
  test.beforeEach(() => {
    test.skip(!runE2E, 'Defina E2E_RUN=1 e suba loja/admin (make dev-up) para executar E2E browser');
  });

  test('admin login page loads', async ({ page }) => {
    await page.goto('/login');
    await expect(page.getByRole('heading', { name: 'Painel administrativo' })).toBeVisible();
  });

  test('store landing loads', async ({ page }) => {
    const storeURL = process.env.E2E_STORE_URL ?? 'http://localhost:5173';
    await page.goto(storeURL);
    await expect(page.locator('body')).toBeVisible();
  });
});
