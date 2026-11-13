import { expect, test } from '@playwright/test';
import { sandbox } from './helpers';

let box: Awaited<ReturnType<typeof sandbox>>;
test.beforeEach(async () => (box = await sandbox()));
test.afterEach(async () => box.dispose());

test('a note dragged onto a folder moves into it', async ({ page }) => {
  await box.api.post('/api/folders', { data: { title: 'Destination', parent_id: box.folder.id } });
  await box.note('# Voyageuse');
  await page.goto(`/folders/${box.folder.id}`);
  const main = page.getByRole('main');

  await main.getByRole('link', { name: /Voyageuse/ }).dragTo(main.getByRole('link', { name: /Destination/ }));
  await expect(main.getByRole('link', { name: /Voyageuse/ })).toBeHidden();
  await main.getByRole('link', { name: /Destination/ }).click();
  await expect(main.getByRole('link', { name: /Voyageuse/ })).toBeVisible();
});

test('a removed note comes back from the trash, then leaves it for good', async ({ page }) => {
  const note = await box.note(`# Recyclée ${box.prefix}`);
  await page.goto(`/folders/${box.folder.id}`);
  const main = page.getByRole('main');

  await main.getByRole('link', { name: /Recyclée/ }).click({ button: 'right' });
  // The open menu hides the page from the roles: wait for the removal itself, or the navigation cancels it.
  const removed = page.waitForResponse((response) => response.url().endsWith(`/api/notes/${note.id}`));
  await page.getByRole('menuitem', { name: 'Supprimer' }).click();
  await removed;
  await expect(main.getByRole('link', { name: /Recyclée/ })).toBeHidden();

  await page.goto('/trash');
  const row = main.getByRole('listitem').filter({ hasText: `Recyclée ${box.prefix}` });
  await row.getByRole('button', { name: 'Restaurer' }).click();
  await expect(row).toBeHidden();
  await page.goto(`/folders/${box.folder.id}`);
  await expect(main.getByRole('link', { name: /Recyclée/ })).toBeVisible();

  await box.api.delete(`/api/notes/${note.id}`);
  await page.goto('/trash');
  await row.getByRole('button', { name: 'Supprimer définitivement' }).click();
  await page.getByRole('dialog').getByRole('button', { name: 'Supprimer' }).click();
  await expect(row).toBeHidden();
  expect((await box.api.post('/api/restore', { data: { ids: [note.id] } })).ok()).toBe(true);
  expect(await (await box.api.get(`/api/notes/${note.id}`)).status()).toBe(404); // Nothing left to restore.
});

// The folder goes to the trash while its new note is open (another device): the note is refused in French, not lost with it (#63).
test('a new note in a trashed folder is refused with a French text', async ({ page }) => {
  await page.goto(`/notes/new?parent=${box.folder.id}`);
  await box.api.delete(`/api/folders/${box.folder.id}`);

  await page.locator('.ProseMirror').click();
  await page.keyboard.type('Perdue ?');
  await expect(page.getByText('Le dossier parent est dans la corbeille').first()).toBeVisible();
  const items = await (await box.api.get(`/api/items?parent=${box.folder.id}`)).json();
  expect(items).toHaveLength(0);
});
