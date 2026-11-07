import { expect, test } from '@playwright/test';
import { API, APP } from '../../playwright.config';

const api = { headers: { Origin: APP } };

test('the account page sets an avatar, shown in the sidebar, then removes it', async ({ page, request }) => {
  try {
    await page.goto('/account');
    // A real 600×400 image (UAvatar shows the initials instead of an image that does not load), cropped and scaled to 256×256 before the upload.
    const wide = await page.evaluate(() => {
      const canvas = Object.assign(document.createElement('canvas'), { width: 600, height: 400 });
      return canvas.toDataURL('image/png').split(',')[1]!;
    });
    await page.locator('input[type=file]').setInputFiles({ name: 'avatar.png', mimeType: 'image/png', buffer: Buffer.from(wide, 'base64') });
    await expect.poll(async () => (await (await request.get(`${API}/api/me`, api)).json()).avatar).toBeTruthy();
    const avatar = page.getByRole('button', { name: 'Menu du compte' }).locator('img');
    await expect(avatar).toHaveAttribute('src', /\/api\/files\//);
    await expect.poll(() => avatar.evaluate((img: HTMLImageElement) => `${img.naturalWidth}×${img.naturalHeight}`)).toBe('256×256');

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
