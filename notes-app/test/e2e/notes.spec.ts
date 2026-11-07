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
  await page.getByRole('option', { name: new RegExp(`^${box.prefix} palette`) }).click();
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

test('# offers the existing tags', async ({ page }) => {
  await box.note(`# ${box.prefix} tagged\n\n#${box.prefix}-tag`);
  const source = await box.note(`# ${box.prefix} source\n\nVoir`);
  await page.goto(`/notes/${source.id}`);
  const editor = page.locator('.ProseMirror');
  await editor.click();
  await page.keyboard.press('ControlOrMeta+End');
  await page.keyboard.type(` #${box.prefix}-t`);
  await page.getByRole('option', { name: `${box.prefix}-tag` }).click();

  await expect.poll(() => box.content(source.id), { timeout: 5000 }).toContain(`#${box.prefix}-tag`);
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

// A 1×1 PNG.
const PNG = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII=';

test('a pasted image is sent to the API, shown, and cited by the note', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} image\n\ntexte`);
  await page.goto(`/notes/${note.id}`);
  const editor = page.locator('.ProseMirror');
  await editor.click();
  await page.keyboard.press('ControlOrMeta+End');
  await editor.evaluate((element, png) => {
    const data = new DataTransfer();
    data.items.add(new File([Uint8Array.from(atob(png), (c) => c.charCodeAt(0))], 'point.png', { type: 'image/png' }));
    element.dispatchEvent(new ClipboardEvent('paste', { clipboardData: data, bubbles: true, cancelable: true }));
  }, PNG);

  const image = editor.locator('img[alt="point.png"]');
  await expect(image).toBeAttached();
  await expect.poll(() => image.evaluate((img: HTMLImageElement) => img.naturalWidth)).toBe(1);
  await expect
    .poll(() => box.content(note.id), { timeout: 5000 })
    .toMatch(/!\[point\.png\]\(\/api\/files\/[0-9a-f-]{36}\)/);
});

test('pinning from the menu lists the note in the sidebar', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} épinglée`);
  await page.goto(`/folders/${box.folder.id}`);
  await page
    .getByRole('main')
    .getByRole('link', { name: new RegExp(`${box.prefix} épinglée`) })
    .click({ button: 'right' });
  await page.getByRole('menuitem', { name: 'Épingler' }).click();
  const pinned = page.locator('section', { has: page.getByRole('heading', { name: 'Épinglées' }) });
  await expect(pinned.getByRole('link', { name: `${box.prefix} épinglée` })).toBeVisible();
  await box.api.patch(`/api/notes/${note.id}`, { data: { pinned: false } });
});

test('N opens a new note in the current folder, the theme button switches to dark', async ({ page }) => {
  await page.goto(`/folders/${box.folder.id}`);
  await expect(page.getByRole('heading', { name: box.prefix })).toBeVisible();
  await page.keyboard.press('n');
  await expect(page).toHaveURL(new RegExp(`/notes/new\\?parent=${box.folder.id}`));

  await page.getByRole('button', { name: 'Thème sombre' }).click();
  await expect(page.locator('html')).toHaveClass(/dark/);
  await page.getByRole('button', { name: 'Thème clair' }).click();
});

test('< > & are saved as typed, not as HTML entities', async ({ page }) => {
  await page.goto(`/notes/new?parent=${box.folder.id}`);
  await page.locator('.ProseMirror').click();
  await page.keyboard.type('a >> b & c < d');
  await expect
    .poll(async () => {
      const items = await (await box.api.get(`/api/items?parent=${box.folder.id}`)).json();
      return items[0] ? box.content(items[0].id) : '';
    })
    .toBe('a >> b & c < d');
});
