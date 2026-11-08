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

// Changed on another device (#72): Reload copies the text typed here, then shows the note as the API has it.
test('after a conflict, reload copies the local text and reads the note again', async ({ page, context }) => {
  await context.grantPermissions(['clipboard-read', 'clipboard-write']);
  const note = await box.note(`# ${box.prefix} conflit\n\nTexte`);
  await page.goto(`/notes/${note.id}`);
  const editor = page.locator('.ProseMirror');
  await expect(editor.getByText('Texte')).toBeVisible();
  await box.api.patch(`/api/notes/${note.id}`, { data: { content: `# ${box.prefix} conflit\n\nAilleurs` } });

  await editor.getByText('Texte').click();
  await page.keyboard.press('End');
  await page.keyboard.type(' ici');
  await page.getByRole('button', { name: 'Recharger' }).click();
  await expect(editor.getByText('Ailleurs')).toBeVisible();
  expect(await page.evaluate(() => navigator.clipboard.readText())).toContain('Texte ici');
  expect(await box.content(note.id)).toContain('Ailleurs');
});

// Past the API's 10 MB, the toast gives the limit in French, not the API's message (#58). The API's answer
// is faked: pasting 10 MB freezes the editor longer than a test.
test('a note too long says its limit', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} longue\n\nTexte`);
  await page.route(`**/api/notes/${note.id}`, (route) =>
    route.request().method() === 'PATCH'
      ? route.fulfill({ status: 413, json: { error: 'request body too large' } })
      : route.fallback(),
  );
  await page.goto(`/notes/${note.id}`);
  await page.locator('.ProseMirror').getByText('Texte').click();
  await page.keyboard.type(' long');
  await expect(page.getByText('La note est trop longue (10 Mo au plus)').first()).toBeVisible();
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

test('a locked note is read-only, in the editor and for the API, until unlocked', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} verrouillée`);
  await page.goto(`/notes/${note.id}`);
  const text = page.locator('.ProseMirror');
  await page.getByRole('button', { name: 'Verrouiller en lecture seule' }).click();
  await expect(text).toHaveAttribute('contenteditable', 'false');
  const refused = await box.api.patch(`/api/notes/${note.id}`, { data: { content: 'autre' } });
  expect(refused.status()).toBe(409);

  // The folder shows its padlock, in the list and on its card.
  await page.goto(`/folders/${box.folder.id}`);
  await expect(page.getByRole('main').getByLabel('Verrouillée')).toBeVisible();
  await expect(page.getByLabel('Verrouillée')).toHaveCount(2); // the folder row and Recent

  await page.getByRole('button', { name: /Rechercher/ }).click();
  await page.getByPlaceholder('Chercher une note, une action…').fill(`${box.prefix} verrouillée`);
  await expect(
    page.getByRole('option', { name: new RegExp(`${box.prefix} verrouillée`) }).getByLabel('Verrouillée'),
  ).toBeVisible();
  await page.keyboard.press('Escape');
  await page.goto(`/search?q=${encodeURIComponent(`${box.prefix} verrouillée`)}`);
  await expect(page.getByRole('main').getByLabel('Verrouillée')).toBeVisible();

  await box.api.patch(`/api/notes/${note.id}`, { data: { pinned: true } });
  await page.reload();
  await expect(page.getByLabel('Verrouillée')).toHaveCount(2); // the search result and Pinned

  await page.goto(`/notes/${note.id}`);
  await expect(text).toHaveAttribute('contenteditable', 'false');
  await page.getByRole('button', { name: 'Verrouiller en lecture seule' }).click();
  await expect(text).toHaveAttribute('contenteditable', 'true');
  await text.press('ControlOrMeta+End');
  await text.pressSequentially(' modifiée');
  await expect.poll(() => box.content(note.id), { timeout: 5000 }).toContain('modifiée');
});

test('a note left open does not overwrite the text written since on another device (#62)', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} deux appareils`);
  await page.goto(`/notes/${note.id}`);
  const text = page.locator('.ProseMirror');
  await expect(text).toContainText('deux appareils');
  await box.api.patch(`/api/notes/${note.id}`, { data: { content: `# ${box.prefix} deux appareils\n\ntéléphone` } });

  await text.press('ControlOrMeta+End');
  await text.pressSequentially(' tablette');
  const toasts = page.getByRole('region', { name: /Notifications/ });
  await expect(toasts.getByText('La note a changé sur un autre appareil', { exact: false })).toBeVisible();
  expect(await box.content(note.id)).toContain('téléphone');
});

test('locking the phone sends the last words at once, and typing on afterwards still saves (#65)', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} verrou`);
  await page.clock.install(); // iOS freezes the 1.5 s timer: here it never runs on its own.
  await page.goto(`/notes/${note.id}`);
  const text = page.locator('.ProseMirror');
  await expect(text).toContainText('verrou');
  const visibility = (state: string) =>
    page.evaluate((state) => {
      Object.defineProperty(document, 'visibilityState', { value: state, configurable: true });
      document.dispatchEvent(new Event('visibilitychange'));
    }, state);

  await text.click();
  await text.press('ControlOrMeta+End');
  await text.pressSequentially(' caché');
  await visibility('hidden');
  await expect.poll(() => box.content(note.id)).toContain('caché');

  await visibility('visible');
  await text.pressSequentially(' revenu');
  await page.clock.runFor(2000);
  await expect.poll(() => box.content(note.id)).toContain('caché revenu');
});

test('a note over 64 KiB still sends its last words when hidden (keepalive limit)', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} longue\n\n${'mot '.repeat(20_000)}fin`);
  await page.clock.install();
  await page.goto(`/notes/${note.id}`);
  const text = page.locator('.ProseMirror');
  await expect(text).toContainText('fin');
  await text.click();
  await text.press('ControlOrMeta+End');
  await text.pressSequentially(' dernier');
  await page.evaluate(() => {
    Object.defineProperty(document, 'visibilityState', { value: 'hidden', configurable: true });
    document.dispatchEvent(new Event('visibilitychange'));
  });
  await expect.poll(() => box.content(note.id)).toContain('dernier');
});

// A full navigation fires pagehide without the router guard; the frozen clock leaves the last words to saveOnClose (#70).
test('closing the page sends the last words of an open note', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} fermée`);
  await page.clock.install();
  await page.goto(`/notes/${note.id}`);
  const text = page.locator('.ProseMirror');
  await expect(text).toContainText('fermée');
  await text.click();
  await text.press('ControlOrMeta+End');
  await text.pressSequentially(' dernier');
  await page.goto('about:blank');
  await expect.poll(() => box.content(note.id)).toContain('dernier');
});

test('closing the page creates the new note typed in it', async ({ page }) => {
  await page.clock.install();
  await page.goto(`/notes/new?parent=${box.folder.id}`);
  await page.locator('.ProseMirror').click();
  await page.keyboard.type(`${box.prefix} nouvelle`);
  await page.goto('about:blank');
  await expect
    .poll(async () => {
      const items: { title: string }[] = await (await box.api.get(`/api/items?parent=${box.folder.id}`)).json();
      return items.map((item) => item.title);
    })
    .toContain(`${box.prefix} nouvelle`);
});

test('closing the page trashes the note emptied in it', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} vidée`);
  await page.clock.install();
  await page.goto(`/notes/${note.id}`);
  const text = page.locator('.ProseMirror');
  await expect(text).toContainText('vidée');
  await text.click();
  await text.press('ControlOrMeta+a');
  await text.press('Backspace');
  await page.goto('about:blank');
  await expect
    .poll(async () => {
      const trashed: { id: string }[] = await (await box.api.get('/api/trash')).json();
      return trashed.some((item) => item.id === note.id);
    })
    .toBe(true);
  await box.api.delete(`/api/trash/${note.id}`);
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

// Pro: the palette renames the open note through its first heading, in the editor and in the source mode.
test('the Pro palette renames the open note', async ({ page }) => {
  const note = await box.note(`intro\n\n## ${box.prefix} old\n\ntext`);
  const rename = async (title: string) => {
    await page.keyboard.press('ControlOrMeta+k');
    await page.getByRole('option', { name: 'Renommer' }).click();
    await page.getByRole('dialog').getByRole('textbox').fill(title);
    await page.getByRole('dialog').getByRole('button', { name: 'Renommer' }).click();
  };
  try {
    await box.api.patch('/api/me', { data: { pro: true } });
    await page.goto(`/notes/${note.id}`);
    await expect(page.locator('.ProseMirror')).toContainText('old');
    await rename(`${box.prefix} new`);
    await expect(page.locator('.ProseMirror h2')).toHaveText(`${box.prefix} new`);
    await expect.poll(() => box.content(note.id), { timeout: 5000 }).toBe(`intro\n\n## ${box.prefix} new\n\ntext`);
    await page.getByRole('button', { name: 'Mode source' }).click();
    await rename(`${box.prefix} source`);
    await expect(page.getByRole('textbox', { name: 'Mode source' })).toHaveValue(
      `intro\n\n## ${box.prefix} source\n\ntext`,
    );
    await expect(page).toHaveTitle(new RegExp(`${box.prefix} source`));
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
    await expect(page.getByText(/\d+ mots?$/)).toBeVisible();
    await page.getByRole('button', { name: 'Mode source' }).click();
    const source = page.getByRole('textbox', { name: 'Mode source' });
    await expect(source).toHaveValue(`# ${box.prefix} source\n\n**gras**`);
    await expect(page.getByText(/\d+ mots?$/)).toBeHidden();
    await source.press('ControlOrMeta+End');
    await source.pressSequentially('\n\n- [ ] tâche');
    await expect.poll(() => box.content(note.id), { timeout: 5000 }).toContain('- [ ] tâche');
    await page.getByRole('button', { name: 'Mode source' }).click();
    await expect(page.locator('.ProseMirror input[type="checkbox"]')).toBeVisible();
  } finally {
    await box.api.patch('/api/me', { data: { pro: false } });
  }
});

// Without the Pro mode the editor still reads and writes tables and callouts: editing a note keeps the ones written elsewhere.
test('without Pro, editing a note keeps its table and callout', async ({ page }) => {
  const markdown = `# ${box.prefix} blocs\n\n| a | b |\n| --- | --- |\n| 1 | 2 |\n\n> [!note]\n> encadré`;
  const note = await box.note(markdown);
  await box.api.patch('/api/me', { data: { pro: false } });
  await page.goto(`/notes/${note.id}`);
  await expect(page.locator('.ProseMirror table')).toBeVisible();
  await page.locator('.ProseMirror h1').click();
  await page.keyboard.press('End');
  await page.keyboard.type(' fin');
  await expect.poll(() => box.content(note.id), { timeout: 5000 }).toContain('blocs fin');
  const saved = await box.content(note.id);
  expect(saved).toContain('> [!note]\n> encadré');
  expect(saved).toMatch(/\| a +\| b +\|/);
});
