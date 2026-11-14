import { expect, test } from '@playwright/test';
import { sandbox } from './helpers';

let box: Awaited<ReturnType<typeof sandbox>>;
test.beforeEach(async () => (box = await sandbox()));
test.afterEach(async () => box.dispose());

// The range spans the two pages read, the first one out of sight by then (#142).
test('a ⇧-click range spans the pages read', async ({ page }) => {
  await box.notes(Array.from({ length: 201 }, (_, i) => `# Page ${String(i).padStart(3, '0')}`));
  await page.goto(`/folders/${box.folder.id}`);
  const main = page.getByRole('main');

  await main.getByRole('link', { name: /Page 000/ }).click({ modifiers: ['ControlOrMeta'] });
  await main.hover();
  await page.mouse.wheel(0, 100_000);
  await main.getByRole('link', { name: /Page 200/ }).click({ modifiers: ['Shift'] });
  await expect(main.getByText('201 sélectionnés')).toBeVisible();
});

test('several items are selected, moved, then removed together', async ({ page }) => {
  const sub = await (
    await box.api.post('/api/folders', { data: { title: 'Rangement', parent_id: box.folder.id } })
  ).json();
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
