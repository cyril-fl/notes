import { expect, test } from '@playwright/test';
import { cursorAtEnd, sandbox } from './helpers';

let box: Awaited<ReturnType<typeof sandbox>>;
test.beforeEach(async () => (box = await sandbox()));
test.afterEach(async () => box.dispose());

test('![[Note]] typed by hand shows the note, resolved by the API (#88)', async ({ page }) => {
  const target = await box.note(`# ${box.prefix} incluse\n\nle contenu inclus`);
  const source = await box.note(`# ${box.prefix} hôte\n\nAvant`);
  await page.goto(`/notes/${source.id}`);
  const editor = page.locator('.ProseMirror').first();
  await cursorAtEnd(editor);
  await page.keyboard.press('Enter');
  await page.keyboard.type(`![[${box.prefix} incluse]]`);
  await page.keyboard.press('Escape');

  await expect
    .poll(() => box.content(source.id), { timeout: 5000 })
    .toContain(`![${box.prefix} incluse](note:${target.id})`);
  await expect(page.locator('.note-embed').getByText('le contenu inclus')).toBeVisible();
});

test('a note that includes itself back stops at one level (#88)', async ({ page }) => {
  const a = await box.note(`# ${box.prefix} A\n\ntexte de A`);
  const b = await box.note(`# ${box.prefix} B\n\ntexte de B\n\n![[${box.prefix} A]]`);
  await box.api.patch(`/api/notes/${a.id}`, {
    data: { content: `# ${box.prefix} A\n\ntexte de A\n\n![[${box.prefix} B]]` },
  });
  await page.goto(`/notes/${a.id}`);

  const embed = page.locator('.note-embed');
  await expect(embed.getByText('texte de B')).toBeVisible();
  // A inside B is a link, not a second copy of A.
  await expect(embed.locator('.note-embed')).toHaveCount(0);
  await expect(embed.locator(`a[data-note="${a.id}"]`)).toBeVisible();
  await embed.locator(`a[data-note="${b.id}"]`).first().click();
  await expect(page).toHaveURL(`/notes/${b.id}`);
});
