import { expect, test } from '@playwright/test';
import { cursorAtEnd, sandbox } from './helpers';

let box: Awaited<ReturnType<typeof sandbox>>;
test.beforeEach(async () => (box = await sandbox()));
test.afterEach(async () => box.dispose());

test('a pasted PDF is sent to the API, cited as a link, and downloads under its name', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} pdf\n\ntexte`);
  await page.goto(`/notes/${note.id}`);
  const editor = page.locator('.ProseMirror');
  await cursorAtEnd(editor);
  await editor.evaluate((element) => {
    const data = new DataTransfer();
    data.items.add(new File(['%PDF-1.7 ...'], 'rapport.pdf', { type: 'application/pdf' }));
    element.dispatchEvent(new ClipboardEvent('paste', { clipboardData: data, bubbles: true, cancelable: true }));
  });

  await expect
    .poll(() => box.content(note.id), { timeout: 5000 })
    .toMatch(/\[rapport\.pdf\]\(\/api\/files\/[0-9a-f-]{36}\)/);

  // Read back from the markdown, the link still downloads.
  await page.reload();
  const download = page.waitForEvent('download');
  await page.locator('.ProseMirror a', { hasText: 'rapport.pdf' }).click();
  expect((await download).suggestedFilename()).toBe('rapport.pdf');
  await expect(page).toHaveURL(`/notes/${note.id}`);
});

// A 1x1 PNG: the API reads the bytes, not the name.
const PNG = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==';

test('the import button of the toolbar sends the picked image and writes it in the note', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} import\n\ntexte`);
  await page.goto(`/notes/${note.id}`);
  await cursorAtEnd(page.locator('.ProseMirror'));

  const chooser = page.waitForEvent('filechooser');
  await page.getByRole('button', { name: 'Importer une image ou un fichier' }).click();
  await (await chooser).setFiles({ name: 'photo.png', mimeType: 'image/png', buffer: Buffer.from(PNG, 'base64') });

  await expect
    .poll(() => box.content(note.id), { timeout: 5000 })
    .toMatch(/!\[photo\.png\]\(\/api\/files\/[0-9a-f-]{36}\)/);

  // Read back from the markdown, the image is shown from the API.
  await page.reload();
  await expect(page.locator('.ProseMirror img')).toHaveAttribute('src', /\/api\/files\/[0-9a-f-]{36}$/);
});

test('a link to a file the API no longer has leaves the note open', async ({ page }) => {
  const note = await box.note(`# ${box.prefix} cassé\n\n[absent.pdf](/api/files/3f2a0000-0000-0000-0000-00000000abcd)`);
  await page.goto(`/notes/${note.id}`);

  // The 404 of the API is JSON: it goes to the thrown-away tab, not over the editor.
  const popup = page.waitForEvent('popup');
  await page.locator('.ProseMirror a', { hasText: 'absent.pdf' }).click();
  await (await popup).close();
  await expect(page).toHaveURL(`/notes/${note.id}`);
  await expect(page.locator('.ProseMirror')).toContainText('absent.pdf');
});
