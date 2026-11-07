import { expect, test } from '@playwright/test';
import { API, APP } from '../../playwright.config';
import { sandbox } from './helpers';

let box: Awaited<ReturnType<typeof sandbox>>;
test.beforeEach(async () => (box = await sandbox()));
test.afterEach(async () => box.dispose());

// The export goes back in as an import: its folders are all taken, so the import is refused and creates nothing.
test('the settings export the notes and import a zip', async ({ page }) => {
  await box.note(`# Exportée ${box.prefix}`);
  await page.goto('/settings');

  const download = page.waitForEvent('download');
  await page.getByRole('link', { name: 'Exporter' }).click();
  const zip = await (await download).path();
  expect(zip).toBeTruthy();

  await page.locator('input[type=file]').setInputFiles(zip);
  await expect(page.getByText('Un dossier du zip existe déjà').first()).toBeVisible();
});

test('a zip is imported into the chosen folder', async ({ page }) => {
  await page.goto('/settings');
  await page.getByRole('button', { name: 'Importer dans' }).click();
  await page.getByRole('option', { name: `/${box.prefix}` }).click();
  await page.locator('input[type=file]').setInputFiles('test/e2e/fixtures/notes.zip');
  await expect(page.getByText('1 note(s) et 1 dossier(s) importés').first()).toBeVisible();

  await page.goto(`/folders/${box.folder.id}`);
  await expect(page.getByRole('main').getByText('Importé')).toBeVisible();
});

// The language is the account's: without the cookie that remembers it, the API gives it back after a reload.
test('the language of the interface is a setting of the account', async ({ page }) => {
  try {
    await page.goto('/settings');
    await page.getByRole('button', { name: 'Langue' }).click();
    await page.getByRole('option', { name: 'English' }).click();
    await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible();

    await page.context().clearCookies({ name: 'notes-locale' });
    await page.reload();
    await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible();
  } finally {
    // The other tests read French.
    await box.api.patch('/api/me', { data: { locale: 'fr' } });
  }
});

// A theme and the trash retention are saved to the account, then set back as they were.
test('the settings save the theme and the trash retention to the account', async ({ page, request }) => {
  const api = { headers: { Origin: APP } };
  const me = async () => (await request.get(`${API}/api/me`, api)).json();
  const before = await me();
  const primary = () => page.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue('--ui-color-primary-500').trim());
  try {
    await page.goto('/settings');
    const notes = await primary();
    await page.getByRole('button', { name: 'Dracula' }).click();
    await expect(page.getByRole('button', { name: 'Dracula' })).toHaveAttribute('aria-pressed', 'true');
    await expect.poll(primary).not.toBe(notes);

    await page.getByRole('button', { name: 'Durée de conservation' }).click();
    await page.getByRole('option', { name: 'Jamais effacés' }).click();

    await expect.poll(me).toMatchObject({ theme: 'dracula', trash_retention_days: 0 });
  } finally {
    await request.patch(`${API}/api/me`, { ...api, data: { theme: before.theme, trash_retention_days: before.trash_retention_days } });
  }
});
