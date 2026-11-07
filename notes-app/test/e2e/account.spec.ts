import { expect, test } from '@playwright/test';

// Starts logged out: without the session the other tests share.
test.use({ storageState: { cookies: [], origins: [] } });

test('without a session the app asks to log in, then goes back where it was', async ({ page }) => {
  await page.goto('/trash');
  await expect(page).toHaveURL(/\/login/);

  await page.getByLabel('Nom').fill(process.env.E2E_NAME ?? 'e2e');
  await page.getByLabel('Mot de passe').fill('wrong password');
  await page.getByRole('button', { name: 'Se connecter' }).click();
  await expect(page.getByRole('alert')).toHaveText('Nom ou mot de passe incorrect.');

  await page.getByLabel('Mot de passe').fill(process.env.E2E_PASSWORD ?? 'e2e-password');
  await page.getByRole('button', { name: 'Se connecter' }).click();
  await expect(page).toHaveURL(/\/trash$/);
  await expect(page.getByRole('heading', { name: 'Corbeille', exact: true })).toBeVisible();

  // Logging out from the account menu of the sidebar leaves the app.
  await page.getByRole('button', { name: 'Menu du compte' }).click();
  await expect(page.getByRole('menu').getByText(process.env.E2E_NAME ?? 'e2e')).toBeVisible();
  await page.getByRole('menuitem', { name: 'Se déconnecter' }).click();
  await expect(page).toHaveURL(/\/login/);
  await page.goto('/');
  await expect(page).toHaveURL(/\/login/);
});
