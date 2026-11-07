import { expect, test } from '@playwright/test';
import { sandbox } from './helpers';

let box: Awaited<ReturnType<typeof sandbox>>;
test.beforeEach(async () => (box = await sandbox()));
test.afterEach(async () => box.dispose());

test('the sidebar collapses and comes back', async ({ page }) => {
  await page.goto('/');
  const search = page.getByRole('button', { name: /Rechercher/ });
  await expect(search).toBeVisible();
  await page.getByRole('button', { name: 'Replier la barre latérale' }).first().click();
  await expect(search).toBeHidden();
  await page.getByRole('button', { name: 'Déployer la barre latérale' }).click();
  await expect(search).toBeVisible();
});

test('a tag is renamed from its menu in the sidebar', async ({ page }) => {
  const tag = box.prefix.replace('-', '');
  await box.note(`# Taguée\n\n#${tag}`);
  await page.goto('/');
  await page.getByRole('link', { name: `#${tag}` }).click({ button: 'right' });
  await page.getByRole('menuitem', { name: 'Renommer' }).click();
  await page.getByRole('dialog').getByRole('textbox').fill(`${tag}x`);
  await page.getByRole('dialog').getByRole('button', { name: 'Renommer' }).click();
  await expect(page.getByRole('link', { name: `#${tag}x` })).toBeVisible();
  await box.api.delete(`/api/tags/${tag}x`); // Trashing the note leaves its tag.
});

test('a folder is made a favorite from its menu', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('navigation').getByText(box.prefix).click({ button: 'right' });
  await page.getByRole('menuitem', { name: 'Ajouter aux favoris' }).click();
  await expect(page.getByRole('navigation').getByRole('link', { name: box.prefix })).toHaveCount(2); // In the tree and in the favorites.
  await expect(page.getByRole('navigation').getByText('Favoris')).toBeVisible();
});

test('on a phone the sidebar opens as a drawer', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 800 });
  await page.goto('/');
  await page.getByRole('button', { name: 'Ouvrir la barre latérale' }).click();
  await expect(page.getByRole('dialog').getByRole('button', { name: /Rechercher/ })).toBeVisible();
});

test('a second folder of the same name is refused', async ({ page }) => {
  await box.api.post('/api/folders', { data: { title: 'Doublon', parent_id: box.folder.id } });
  await page.goto(`/folders/${box.folder.id}`);
  await page.getByRole('main').getByRole('button', { name: 'Nouveau dossier' }).click();
  await page.getByRole('dialog').getByRole('textbox').fill('doublon');
  await page.getByRole('dialog').getByRole('button', { name: 'Créer' }).click();
  await expect(page.getByText('Un dossier porte déjà ce nom à cet endroit.')).toBeVisible();
});
