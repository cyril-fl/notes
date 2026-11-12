import { expect, test, type Route } from '@playwright/test';
import { cursorAtEnd, sandbox, sourceText } from './helpers';

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
  await cursorAtEnd(editor);
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
  await cursorAtEnd(editor);
  await page.keyboard.type(` #${box.prefix}-t`);
  await page.getByRole('option', { name: `${box.prefix}-tag` }).click();

  await expect.poll(() => box.content(source.id), { timeout: 5000 }).toContain(`#${box.prefix}-tag`);
});

test('a [[link]] typed by hand waits, then the API points it to its note', async ({ page }) => {
  const target = await box.note(`# ${box.prefix} main`);
  const source = await box.note(`# ${box.prefix} tapé\n\nVoir`);
  await page.goto(`/notes/${source.id}`);
  const editor = page.locator('.ProseMirror');
  await cursorAtEnd(editor);
  await page.keyboard.type(` [[${box.prefix} main]]`);
  await page.keyboard.press('Escape');
  await expect(editor.locator('a.wikilink-waiting')).toBeVisible();
  await expect.poll(() => box.content(source.id), { timeout: 5000 }).toContain(`](note:${target.id})`);
});

test('a [[link]] several notes answer looks ambiguous, a link to none only waits', async ({ page }) => {
  await box.note(`# ${box.prefix} twin`);
  await box.note(`# ${box.prefix} twin`);
  const source = await box.note(`# ${box.prefix} ambigu\n\n[[${box.prefix} twin]] [[${box.prefix} nowhere]]`);
  await page.goto(`/notes/${source.id}`);
  const editor = page.locator('.ProseMirror');
  await expect(editor.locator('a.wikilink-ambiguous')).toHaveText(`${box.prefix} twin`);
  await expect(editor.locator('a.wikilink-waiting:not(.wikilink-ambiguous)')).toHaveText(`${box.prefix} nowhere`);
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
  await cursorAtEnd(editor);
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
  await cursorAtEnd(editor, 'Écrire à');
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

// Changed on another device (#78): changes in different places merge on their own, both kept.
test('after a conflict, changes in different places merge on their own', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} fusion\n\nTexte\n\nFin`);
  await page.goto(`/notes/${note.id}`);
  const editor = page.locator('.ProseMirror');
  await expect(editor.getByText('Texte')).toBeVisible();
  await box.api.patch(`/api/notes/${note.id}`, {
    data: { content: `# ${box.prefix} fusion\n\nTexte\n\nFin ailleurs` },
  });

  await cursorAtEnd(editor, 'Texte');
  await page.keyboard.type(' ici');
  await expect(page.getByText('les deux textes sont fusionnés').last()).toBeVisible();
  await expect(editor.getByText('Fin ailleurs')).toBeVisible();
  await expect
    .poll(() => box.content(note.id), { timeout: 5000 })
    .toBe(`# ${box.prefix} fusion\n\nTexte ici\n\nFin ailleurs`);
});

// The same line changed on both sides (#78): the dialog asks, the text typed here kept in the history as a conflict.
test('after a conflict in the same place, the dialog asks which text to keep', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} conflit\n\nTexte`);
  await page.goto(`/notes/${note.id}`);
  const editor = page.locator('.ProseMirror');
  await expect(editor.getByText('Texte')).toBeVisible();
  await box.api.patch(`/api/notes/${note.id}`, { data: { content: `# ${box.prefix} conflit\n\nAilleurs` } });

  await cursorAtEnd(editor, 'Texte');
  await page.keyboard.type(' ici');
  const dialog = page.getByRole('dialog');
  await expect(dialog.getByRole('heading', { name: 'Conflits (1 à résoudre)' })).toBeVisible();
  await expect(dialog.getByRole('region', { name: 'Conflit 1' })).toContainText('Ailleurs');
  await dialog.getByRole('button', { name: 'Les deux' }).click();
  await dialog.getByRole('button', { name: 'Enregistrer' }).click();
  await expect
    .poll(() => box.content(note.id), { timeout: 5000 })
    .toBe(`# ${box.prefix} conflit\n\nTexte ici\n\nAilleurs`);
  const versions: { id: string; conflict: boolean }[] = await (
    await box.api.get(`/api/notes/${note.id}/versions`)
  ).json();
  const kept = versions.find((version) => version.conflict);
  const version = await (await box.api.get(`/api/notes/${note.id}/versions/${kept?.id}`)).json();
  expect(version.content).toContain('Texte ici');
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
  await cursorAtEnd(editor);
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

test('the size of an image is picked over it and kept as the title of its markdown', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} taille\n\n![chat](https://x.org/chat.png "small")`);
  await page.goto(`/notes/${note.id}`);
  const image = page.locator('.ProseMirror img[alt="chat"]');
  await expect(image).toHaveClass(/note-image-small/);

  await image.click();
  await page.getByRole('button', { name: 'Moyenne' }).click();
  await expect(image).toHaveClass(/note-image-medium/);
  await expect
    .poll(() => box.content(note.id), { timeout: 5000 })
    .toContain('![chat](https://x.org/chat.png "medium")');

  await page.getByRole('button', { name: 'Grande' }).click();
  await expect(image).toHaveClass(/note-image-large/);
  await expect.poll(() => box.content(note.id), { timeout: 5000 }).toContain('![chat](https://x.org/chat.png "large")');

  await page.getByRole('button', { name: "Taille d'origine" }).click();
  await expect(image).not.toHaveClass(/note-image/);
  await expect.poll(() => box.content(note.id), { timeout: 5000 }).toContain('![chat](https://x.org/chat.png)');
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
  await cursorAtEnd(text);
  await text.pressSequentially(' modifiée');
  await expect.poll(() => box.content(note.id), { timeout: 5000 }).toContain('modifiée');
});

test('a note left open does not overwrite the text written since on another device (#62)', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} deux appareils\n\nTexte`);
  await page.goto(`/notes/${note.id}`);
  const text = page.locator('.ProseMirror');
  await expect(text).toContainText('deux appareils');
  await box.api.patch(`/api/notes/${note.id}`, {
    data: { content: `# ${box.prefix} deux appareils\n\nTexte\n\ntéléphone` },
  });

  await cursorAtEnd(text, 'deux appareils'); // The title, a line away from the phone's paragraph: no conflict.
  await text.pressSequentially(' tablette');
  const toasts = page.getByRole('region', { name: /Notifications/ });
  await expect(toasts.getByText('les deux textes sont fusionnés', { exact: false })).toBeVisible();
  await expect.poll(() => box.content(note.id), { timeout: 5000 }).toContain('tablette');
  expect(await box.content(note.id)).toContain('téléphone');
});

// Locked on the other device, nothing merges: the toast offers to reload, the note keeps the other device's text.
test('a note locked since on another device is not overwritten, the toast offers to reload', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} verrouillée ailleurs`);
  await page.goto(`/notes/${note.id}`);
  const text = page.locator('.ProseMirror');
  await expect(text).toContainText('verrouillée ailleurs');
  await box.api.patch(`/api/notes/${note.id}`, {
    data: { content: `# ${box.prefix} verrouillée ailleurs\n\ntéléphone` },
  });
  await box.api.patch(`/api/notes/${note.id}`, { data: { readonly: true } });

  await cursorAtEnd(text);
  await text.pressSequentially(' tablette');
  const toasts = page.getByRole('region', { name: /Notifications/ });
  await expect(toasts.getByText('La note a changé sur un autre appareil', { exact: false })).toBeVisible();
  await expect(toasts.getByRole('button', { name: 'Recharger' })).toBeVisible();
  const content = await box.content(note.id);
  expect(content).toContain('téléphone');
  expect(content).not.toContain('tablette');
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

  await cursorAtEnd(text);
  await text.pressSequentially(' caché');
  await visibility('hidden');
  await expect.poll(() => box.content(note.id)).toContain('caché');

  await visibility('visible');
  await text.pressSequentially(' revenu');
  await page.clock.runFor(2000);
  await expect.poll(() => box.content(note.id)).toContain('caché revenu');
});

test('the last words that fail to go when the phone locks are sent again once it is back (#90)', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} panne`);
  await page.clock.install();
  await page.goto(`/notes/${note.id}`);
  const text = page.locator('.ProseMirror');
  await expect(text).toContainText('panne');
  const visibility = (state: string) =>
    page.evaluate((state) => {
      Object.defineProperty(document, 'visibilityState', { value: state, configurable: true });
      document.dispatchEvent(new Event('visibilitychange'));
    }, state);

  await cursorAtEnd(text);
  await text.pressSequentially(' perdu');
  const offline = (route: Route) => (route.request().method() === 'PATCH' ? route.abort() : route.fallback());
  await page.route('**/api/notes/*', offline);
  await visibility('hidden');
  await page.waitForTimeout(500);
  expect(await box.content(note.id)).not.toContain('perdu');

  await page.unroute('**/api/notes/*', offline);
  await visibility('visible');
  await page.clock.runFor(2000);
  await expect.poll(() => box.content(note.id)).toContain('perdu');
});

test('a note over 64 KiB still sends its last words when hidden (keepalive limit)', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} longue\n\n${'mot '.repeat(20_000)}fin`);
  await page.clock.install();
  await page.goto(`/notes/${note.id}`);
  const text = page.locator('.ProseMirror');
  await expect(text).toContainText('fin');
  await cursorAtEnd(text);
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
  await cursorAtEnd(text);
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
    await box.api.patch('/api/me', { data: { plan: 'pro' } });
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
    await box.api.patch('/api/me', { data: { plan: 'base' } });
  }
});

// Pro: the palette renames the open note through its first heading, in the editor and in the source mode, where the word count is hidden.
test('the Pro palette renames the open note', async ({ page }) => {
  const note = await box.note(`intro\n\n## ${box.prefix} old\n\ntext`);
  const rename = async (title: string) => {
    await page.keyboard.press('ControlOrMeta+k');
    await page.getByRole('option', { name: 'Renommer' }).click();
    await page.getByRole('dialog').getByRole('textbox').fill(title);
    await page.getByRole('dialog').getByRole('button', { name: 'Renommer' }).click();
  };
  try {
    await box.api.patch('/api/me', { data: { plan: 'pro' } });
    await page.goto(`/notes/${note.id}`);
    await expect(page.locator('.ProseMirror')).toContainText('old');
    await rename(`${box.prefix} new`);
    await expect(page.locator('.ProseMirror h2')).toHaveText(`${box.prefix} new`);
    await expect.poll(() => box.content(note.id), { timeout: 5000 }).toBe(`intro\n\n## ${box.prefix} new\n\ntext`);
    // The Pro word count belongs to the editor: hidden in the source mode.
    await expect(page.getByText(/\d+ mots?$/)).toBeVisible();
    await page.getByRole('button', { name: 'Mode source' }).click();
    await expect(page.getByText(/\d+ mots?$/)).toBeHidden();
    await rename(`${box.prefix} source`);
    await expect
      .poll(() => sourceText(page.getByRole('textbox', { name: 'Mode source' })))
      .toBe(`intro\n\n## ${box.prefix} source\n\ntext`);
    await expect(page).toHaveTitle(new RegExp(`${box.prefix} source`));
  } finally {
    await box.api.patch('/api/me', { data: { plan: 'base' } });
  }
});

// The source mode, for everyone: the markdown as text, saves what is typed there, and the editor starts from it.
test('the source mode edits the markdown', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} source\n\n**gras**`);
  await page.goto(`/notes/${note.id}`);
  await expect(page.locator('.ProseMirror')).toContainText('gras');
  await page.getByRole('button', { name: 'Mode source' }).click();
  const source = page.getByRole('textbox', { name: 'Mode source' });
  await expect.poll(() => sourceText(source)).toBe(`# ${box.prefix} source\n\n**gras**`);
  await source.fill(`# ${box.prefix} source\n\n**gras**\n\n- [ ] tâche`);
  await expect.poll(() => box.content(note.id), { timeout: 5000 }).toContain('- [ ] tâche');
  await page.getByRole('button', { name: 'Mode source' }).click();
  await expect(page.locator('.ProseMirror input[type="checkbox"]')).toBeVisible();
});

// The reason the source mode is a code editor (#112): one cursor per occurrence, which the rich editor cannot carry.
test('Cmd+D adds a cursor on the next occurrence of the source mode', async ({ page }) => {
  const note = await box.note('foo bar\n\nfoo baz\n\nfoo qux');
  await page.goto(`/notes/${note.id}`);
  await expect(page.locator('.ProseMirror')).toContainText('foo bar');
  await page.getByRole('button', { name: 'Mode source' }).click();
  const source = page.getByRole('textbox', { name: 'Mode source' });
  await source.click();
  await page.keyboard.press('ControlOrMeta+Home');
  for (let i = 0; i < 3; i++) await page.keyboard.press('Shift+ArrowRight'); // The first foo selected.
  await page.keyboard.press('ControlOrMeta+d');
  await page.keyboard.press('Control+d'); // The way out of Safari, which keeps Cmd+D for its bookmarks.
  await page.keyboard.type('xxx');
  await expect.poll(() => sourceText(source)).toBe('xxx bar\n\nxxx baz\n\nxxx qux');
  await expect.poll(() => box.content(note.id), { timeout: 5000 }).toBe('xxx bar\n\nxxx baz\n\nxxx qux');
});

test('the empty Pro context panel explains the related notes and the history', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} history empty\n\nfirst`);
  try {
    await box.api.patch('/api/me', { data: { plan: 'pro' } });
    await page.goto(`/notes/${note.id}`);
    await expect(page.getByText("Aucune note liée pour l'instant.")).toBeVisible();
    const panel = page.getByRole('region', { name: 'Historique des versions' });
    await expect(panel.getByText("Aucune version antérieure n'est encore disponible.")).toBeVisible();
    await expect(panel.getByText('Une copie est enregistrée après chaque modification.')).toBeVisible();
    await expect(
      panel.getByText('Une nouvelle version après 15 minutes de pause, et au moins une par heure.'),
    ).toBeVisible();
  } finally {
    await box.api.patch('/api/me', { data: { plan: 'base' } });
  }
});

// Pro (#61): the text replaced by an edit is kept, and restoring it saves it back.
test('the Pro history restores an earlier version', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} history\n\nfirst`);
  await box.api.patch(`/api/notes/${note.id}`, { data: { content: `# ${box.prefix} history\n\nsecond` } });
  try {
    await box.api.patch('/api/me', { data: { plan: 'pro' } });
    await page.goto(`/notes/${note.id}`);
    await expect(page.locator('.ProseMirror')).toContainText('second');
    // The context panel lists the versions on a wide screen: a click shows one.
    const panel = page.getByRole('region', { name: 'Historique des versions' });
    await panel
      .getByRole('button', { name: /\d{2}:\d{2}/ })
      .first()
      .click();
    await expect(page.getByRole('dialog')).toContainText('first');
    // Compared with the text as saved (#78): the version's line struck, the text's underlined.
    await page.getByRole('button', { name: 'Comparer au texte actuel' }).click();
    await expect(page.getByRole('dialog').locator('del')).toContainText('first');
    await expect(page.getByRole('dialog').locator('ins')).toContainText('second');
    await page.getByRole('button', { name: 'Restaurer cette version' }).click();
    await expect(page.locator('.ProseMirror')).toContainText('first');
    await expect.poll(() => box.content(note.id), { timeout: 5000 }).toBe(`# ${box.prefix} history\n\nfirst`);
  } finally {
    await box.api.patch('/api/me', { data: { plan: 'base' } });
  }
});

// A fork: the version becomes a new note in the same folder, the note itself left as it is.
test('the Pro history forks an earlier version into a new note', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} fork\n\nfirst #${box.prefix}-old`);
  await box.api.patch(`/api/notes/${note.id}`, {
    data: { content: `# ${box.prefix} fork\n\nsecond #${box.prefix}-new` },
  });
  try {
    await box.api.patch('/api/me', { data: { plan: 'pro' } });
    await page.goto(`/notes/${note.id}`);
    await page
      .getByRole('region', { name: 'Historique des versions' })
      .getByRole('button', { name: /\d{2}:\d{2}/ })
      .first()
      .click();
    await expect(page.getByRole('dialog')).toContainText('first');
    await page.getByRole('button', { name: 'Nouvelle note depuis cette version' }).click();
    await expect(page).not.toHaveURL(new RegExp(note.id));
    await expect(page.locator('.ProseMirror')).toContainText(`${box.prefix} fork (version du`);
    await expect(page.locator('.ProseMirror')).toContainText('first');
    // The copy carries the tags of that version, not those the note wears today (#115).
    const copy = page.url().split('/').pop()!;
    const copied = await box.content(copy);
    expect(copied).toContain(`#${box.prefix}-old`);
    expect(copied).not.toContain(`#${box.prefix}-new`);
    expect(await box.content(note.id)).toBe(`# ${box.prefix} fork\n\nsecond #${box.prefix}-new`);
  } finally {
    await box.api.patch('/api/me', { data: { plan: 'base' } });
  }
});

// Without the Pro mode the editor still reads and writes tables and callouts: editing a note keeps the ones written elsewhere.
test('without Pro, editing a note keeps its table and callout', async ({ page }) => {
  const markdown = `# ${box.prefix} blocs\n\n| a | b |\n| --- | --- |\n| 1 | 2 |\n\n> [!note]\n> encadré`;
  const note = await box.note(markdown);
  await box.api.patch('/api/me', { data: { plan: 'base' } });
  await page.goto(`/notes/${note.id}`);
  await expect(page.locator('.ProseMirror table')).toBeVisible();
  await cursorAtEnd(page.locator('.ProseMirror'), `${box.prefix} blocs`);
  await page.keyboard.type(' fin');
  await expect.poll(() => box.content(note.id), { timeout: 5000 }).toContain('blocs fin');
  const saved = await box.content(note.id);
  expect(saved).toContain('> [!note]\n> encadré');
  expect(saved).toMatch(/\| a +\| b +\|/);
});

test.describe('on a touch screen', () => {
  test.use({ hasTouch: true });

  test('a toolbar button selects the whole note', async ({ page }) => {
    const note = await box.note(`# ${box.prefix} tout\n\nun\n\ndeux`);
    await page.goto(`/notes/${note.id}`);
    await page.getByRole('button', { name: 'Tout sélectionner' }).click();
    await expect.poll(() => page.evaluate(() => getSelection()?.toString())).toMatch(/tout[\s\S]*un[\s\S]*deux/);
  });

  test('toolbar buttons undo and redo the typing', async ({ page }) => {
    const note = await box.note(`# ${box.prefix} annuler`);
    await page.goto(`/notes/${note.id}`);
    await cursorAtEnd(page.locator('.ProseMirror'), `${box.prefix} annuler`);
    await page.keyboard.type(' faute');
    await page.getByRole('button', { name: 'Annuler' }).click();
    await expect(page.locator('.ProseMirror')).not.toContainText('faute');
    await page.getByRole('button', { name: 'Rétablir' }).click();
    await expect(page.locator('.ProseMirror')).toContainText('faute');
  });

  test('a toolbar button empties the note once confirmed', async ({ page }) => {
    const note = await box.note(`# ${box.prefix} vider\n\ntexte`);
    await page.goto(`/notes/${note.id}`);
    await page.getByRole('button', { name: 'Vider la note' }).click();
    await page.getByRole('dialog').getByRole('button', { name: 'Vider la note' }).click();
    await expect(page.locator('.ProseMirror')).not.toContainText('texte');
    // An empty note is not saved: what is typed next replaces the text.
    await page.keyboard.type('neuf');
    await expect.poll(() => box.content(note.id), { timeout: 5000 }).toBe('neuf');
  });
});
