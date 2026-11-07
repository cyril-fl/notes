import { expect, test } from '@playwright/test';
import { sandbox } from './helpers';

let box: Awaited<ReturnType<typeof sandbox>>;
test.beforeEach(async () => (box = await sandbox()));
test.afterEach(async () => box.dispose());

test('several items are selected, moved, then removed together', async ({ page }) => {
  const sub = await (await box.api.post('/api/folders', { data: { title: 'Rangement', parent_id: box.folder.id } })).json();
  await box.note('# Alpha');
  await box.note('# Bravo');
  await page.goto(`/folders/${box.folder.id}`);
  const main = page.getByRole('main');

  await main.getByRole('link', { name: /Alpha/ }).click({ modifiers: ['ControlOrMeta'] });
  await main.getByRole('link', { name: /Bravo/ }).click({ modifiers: ['ControlOrMeta'] });
  await expect(main.getByText('2 sélectionnés')).toBeVisible();
  await expect(page).toHaveURL(new RegExp(box.folder.id)); // ⌘-click selects, it does not open.

  await main.getByRole('button', { name: 'Déplacer vers' }).click();
  await page.getByRole('option', { name: `/${box.prefix}/Rangement` }).click();
  await expect(main.getByRole('link', { name: /Alpha/ })).toBeHidden();

  await page.goto(`/folders/${sub.id}`);
  await main.getByRole('link', { name: /Alpha/ }).click({ modifiers: ['ControlOrMeta'] });
  await main.getByRole('link', { name: /Bravo/ }).click({ modifiers: ['Shift'] });
  await main.getByRole('button', { name: 'Supprimer' }).click();
  await expect(main.getByText('Ce dossier est vide')).toBeVisible();
});
