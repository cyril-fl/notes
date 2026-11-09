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
