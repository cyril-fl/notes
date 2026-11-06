import { expect, test } from '@playwright/test';
import { sandbox } from './helpers';

let box: Awaited<ReturnType<typeof sandbox>>;
test.beforeEach(async () => (box = await sandbox()));
test.afterEach(async () => box.dispose());

test('the palette finds a note and opens it', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} palette\n\ntexte`);
  await page.goto('/');
  await expect(page.getByRole('button', { name: /Rechercher/ })).toBeVisible();
  await page.keyboard.press('ControlOrMeta+k');
  await page.getByPlaceholder('Chercher une note, une action…').fill(`${box.prefix} palette`);
  await page.getByRole('option', { name: new RegExp(`${box.prefix} palette`) }).first().click();
  await expect(page).toHaveURL(`/notes/${note.id}`);
});

test('[[ offers the notes and links by id', async ({ page }) => {
  const target = await box.note(`# ${box.prefix} cible`);
  const source = await box.note(`# ${box.prefix} source\n\nVoir`);
  await page.goto(`/notes/${source.id}`);
  const editor = page.locator('.ProseMirror');
  await editor.click();
  await page.keyboard.press('ControlOrMeta+End');
  await page.keyboard.type(' [[cible');
  await page.getByRole('option', { name: new RegExp(`${box.prefix} cible`) }).click();
  await expect(editor.locator(`a[data-note="${target.id}"]`)).toBeVisible();

  await expect.poll(() => box.content(source.id), { timeout: 5000 }).toContain(`](note:${target.id})`);
  await editor.locator(`a[data-note="${target.id}"]`).click();
  await expect(page).toHaveURL(`/notes/${target.id}`);
});

test('a [[link]] typed by hand waits, then the API points it to its note', async ({ page }) => {
  const target = await box.note(`# ${box.prefix} main`);
  const source = await box.note(`# ${box.prefix} tapé\n\nVoir`);
  await page.goto(`/notes/${source.id}`);
  const editor = page.locator('.ProseMirror');
  await editor.click();
  await page.keyboard.press('ControlOrMeta+End');
  await page.keyboard.type(` [[${box.prefix} main]]`);
  await page.keyboard.press('Escape');
  await expect(editor.locator('a.wikilink-waiting')).toBeVisible();
  await expect.poll(() => box.content(source.id), { timeout: 5000 }).toContain(`](note:${target.id})`);
});

test('pinning from the menu lists the note in the sidebar', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} épinglée`);
  await page.goto(`/folders/${box.folder.id}`);
  await page.getByRole('main').getByRole('link', { name: new RegExp(`${box.prefix} épinglée`) }).click({ button: 'right' });
  await page.getByRole('menuitem', { name: 'Épingler' }).click();
  const pinned = page.locator('section', { has: page.getByRole('heading', { name: 'Épinglées' }) });
  await expect(pinned.getByRole('link', { name: `${box.prefix} épinglée` })).toBeVisible();
  await box.api.patch(`/api/notes/${note.id}`, { data: { pinned: false } });
});

test('N opens a new note in the current folder, the theme button switches to dark', async ({ page }) => {
  await page.goto(`/folders/${box.folder.id}`);
  await page.locator('body').click();
  await page.keyboard.press('n');
  await expect(page).toHaveURL(new RegExp(`/notes/new\\?parent=${box.folder.id}`));

  await page.getByRole('button', { name: 'Thème sombre' }).click();
  await expect(page.locator('html')).toHaveClass(/dark/);
  await page.getByRole('button', { name: 'Thème clair' }).click();
});
