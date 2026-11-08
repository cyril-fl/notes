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

test('a web link turns into a large card, then a mini one, kept in the markdown', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} carte\n\nVoir [Exemple](https://example.com) ici`);
  await page.goto(`/notes/${note.id}`);
  const editor = page.locator('.ProseMirror');
  await editor.getByText('Exemple').click();
  await page.getByRole('button', { name: 'Large' }).click();
  await expect(editor.locator('a.link-card-large')).toContainText('Exemple');
  await expect.poll(() => box.content(note.id), { timeout: 5000 }).toContain('[Exemple](https://example.com "large")');

  await editor.locator('a.link-card').click();
  await page.getByRole('button', { name: 'Mini' }).click();
  await expect.poll(() => box.content(note.id), { timeout: 5000 }).toContain('[Exemple](https://example.com "mini")');
  await page.reload();
  await expect(editor.locator('a.link-card-mini')).toContainText('Exemple');

  await editor.locator('a.link-card').click();
  await page.getByRole('button', { name: 'Hyperlien' }).click();
  await expect.poll(() => box.content(note.id), { timeout: 5000 }).toContain('[Exemple](https://example.com) ici');
});

test('a pasted address becomes a mini card', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} collé\n\nVoir`);
  await page.goto(`/notes/${note.id}`);
  const editor = page.locator('.ProseMirror');
  await editor.click();
  await page.keyboard.press('ControlOrMeta+End');
  await editor.evaluate((el) => {
    const data = new DataTransfer();
    data.setData('text/plain', 'https://example.com/page');
    el.dispatchEvent(new ClipboardEvent('paste', { clipboardData: data, bubbles: true, cancelable: true }));
  });
  await expect(editor.locator('a.link-card-mini')).toContainText('example.com');
  await expect
    .poll(() => box.content(note.id), { timeout: 5000 })
    .toContain('[example.com](https://example.com/page "mini")');
});

test('an email is a link: typed, pasted or given to the link button', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} mail\n\nÉcrire à`);
  await page.goto(`/notes/${note.id}`);
  const editor = page.locator('.ProseMirror');
  await editor.getByText('Écrire à').click();
  await page.keyboard.press('End');
  await page.keyboard.type(' tape@example.com ');
  await expect(editor.locator('a[href="mailto:tape@example.com"]')).toBeVisible();

  await editor.evaluate((el) => {
    const data = new DataTransfer();
    data.setData('text/plain', 'colle@example.com');
    el.dispatchEvent(new ClipboardEvent('paste', { clipboardData: data, bubbles: true, cancelable: true }));
  });
  await expect(editor.locator('a.link-card-mini')).toContainText('colle@example.com');

  await page.getByRole('button', { name: 'Lien', exact: true }).click();
  await page.getByPlaceholder('https://… ou une adresse mail').fill('bouton@example.com');
  await page.getByRole('button', { name: 'Ajouter' }).click();
  await expect(editor.locator('a.link-card-mini').nth(1)).toContainText('bouton@example.com');
  await expect
    .poll(() => box.content(note.id), { timeout: 5000 })
    .toContain('[bouton@example.com](mailto:bouton@example.com "mini")');
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

// In system mode the OS decides: the sidebar has no theme button. Set back to light, as every run expects.
test('N opens a new note in the current folder, the theme button switches to dark', async ({ page }) => {
  try {
    await box.api.patch('/api/me', { data: { color_mode: 'system' } });
    await page.goto(`/folders/${box.folder.id}`);
    await expect(page.getByRole('heading', { name: box.prefix })).toBeVisible();
    await expect(page.getByRole('button', { name: /Thème (sombre|clair)/ })).toHaveCount(0);

    await box.api.patch('/api/me', { data: { color_mode: 'light' } });
    await page.reload();
    await expect(page.getByRole('heading', { name: box.prefix })).toBeVisible();
    await page.keyboard.press('n');
    await expect(page).toHaveURL(new RegExp(`/notes/new\\?parent=${box.folder.id}`));

    await page.getByRole('button', { name: 'Thème sombre' }).click();
    await expect(page.locator('html')).toHaveClass(/dark/);
    await page.getByRole('button', { name: 'Thème clair' }).click();
  } finally {
    await box.api.patch('/api/me', { data: { color_mode: 'light' } });
  }
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

// Pro: the palette moves the open note, and its path under the text follows.
test('the Pro palette moves the open note', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} moved`);
  const sub = await (await box.api.post('/api/folders', { data: { title: 'sub', parent_id: box.folder.id } })).json();
  try {
    await box.api.patch('/api/me', { data: { pro: true } });
    await page.goto(`/notes/${note.id}`);
    await expect(page.locator('.ProseMirror')).toContainText('moved');
    await page.keyboard.press('ControlOrMeta+k');
    await page.getByRole('option', { name: 'Déplacer vers…' }).click();
    await page.getByPlaceholder('Chercher une note, une action…').fill(`${box.prefix}/sub`);
    await page.getByRole('option', { name: `/${box.prefix}/sub` }).click();
    await expect(
      page.getByRole('navigation', { name: "Fil d'Ariane" }).getByRole('link', { name: 'sub' }),
    ).toBeVisible();
    const moved = await (await box.api.get(`/api/notes/${note.id}`)).json();
    expect(moved.item.parent_id).toBe(sub.id);
  } finally {
    await box.api.patch('/api/me', { data: { pro: false } });
  }
});

// Pro: the source mode shows the markdown as text, saves what is typed there, and the editor starts from it.
test('the Pro source mode edits the markdown', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} source\n\n**gras**`);
  try {
    await box.api.patch('/api/me', { data: { pro: true } });
    await page.goto(`/notes/${note.id}`);
    await expect(page.locator('.ProseMirror')).toContainText('gras');
    await page.getByRole('button', { name: 'Mode source' }).click();
    const source = page.getByRole('textbox', { name: 'Mode source' });
    await expect(source).toHaveValue(`# ${box.prefix} source\n\n**gras**`);
    await source.press('ControlOrMeta+End');
    await source.pressSequentially('\n\n- [ ] tâche');
    await expect.poll(() => box.content(note.id), { timeout: 5000 }).toContain('- [ ] tâche');
    await page.getByRole('button', { name: 'Mode source' }).click();
    await expect(page.locator('.ProseMirror input[type="checkbox"]')).toBeVisible();
  } finally {
    await box.api.patch('/api/me', { data: { pro: false } });
  }
});
