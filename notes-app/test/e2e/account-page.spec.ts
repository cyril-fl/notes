import { expect, test } from '@playwright/test';
import { API, APP } from '../../playwright.config';

const api = { headers: { Origin: APP } };
// A real PNG of one pixel: UAvatar shows the initials instead of an image that does not load.
const png = {
  name: 'avatar.png',
  mimeType: 'image/png',
  buffer: Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII=', 'base64'),
};

test('the account page sets an avatar, shown in the sidebar, then removes it', async ({ page, request }) => {
  try {
    await page.goto('/account');
    await page.locator('input[type=file]').setInputFiles(png);
    await expect.poll(async () => (await (await request.get(`${API}/api/me`, api)).json()).avatar).toBeTruthy();
    await expect(page.getByRole('button', { name: 'Menu du compte' }).locator('img')).toHaveAttribute('src', /\/api\/files\//);

    await page.getByRole('button', { name: 'Retirer' }).click();
    await expect(page.getByRole('button', { name: 'Menu du compte' }).locator('img')).toHaveCount(0);
  } finally {
    await request.patch(`${API}/api/me`, { ...api, data: { avatar: '' } });
  }
});

// The name logs in: it is set back for the next runs.
test('the account page renames the account', async ({ page, request }) => {
  const before = (await (await request.get(`${API}/api/me`, api)).json()).name;
  try {
    await page.goto('/account');
    await page.getByLabel('Nom').fill(`${before}-renamed`);
    await page.getByRole('button', { name: 'Enregistrer' }).click();
    await expect(page.getByText('Nom enregistré').first()).toBeVisible();
    await expect(page.getByText(`Connecté en tant que ${before}-renamed`)).toBeVisible();
  } finally {
    await request.patch(`${API}/api/me`, { ...api, data: { name: before } });
  }
});

// A real change would log out the session the other tests share: only the wrong password is tried.
test('the account page refuses a wrong current password', async ({ page }) => {
  await page.goto('/account');
  await page.getByLabel('Mot de passe actuel').fill('wrong password');
  await page.getByLabel('Nouveau mot de passe').fill('another password');
  await page.getByRole('button', { name: 'Changer le mot de passe' }).click();
  await expect(page.getByText('Le mot de passe actuel est faux.').first()).toBeVisible();
});
