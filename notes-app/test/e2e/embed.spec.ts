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
  // Nothing of the typed brackets is left around it.
  expect(await box.content(source.id)).toContain(`Avant\n\n![${box.prefix} incluse](note:${target.id})`);
  expect(await box.content(source.id)).not.toContain('\\[');
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

test('![[ offers the notes and embeds the one picked; in a sentence it is a link (#88)', async ({ page }) => {
  const target = await box.note(`# ${box.prefix} choisie\n\nson texte`);
  const source = await box.note(`# ${box.prefix} hôte\n\nAvant`);
  await page.goto(`/notes/${source.id}`);
  const editor = page.locator('.ProseMirror').first();
  await cursorAtEnd(editor);
  await page.keyboard.press('Enter');
  await page.keyboard.type(`![[${box.prefix} chois`, { delay: 50 });
  await page.getByRole('option', { name: new RegExp(`${box.prefix} choisie`) }).click();

  const embed = page.locator('.note-embed');
  await expect(embed.getByText('son texte')).toBeVisible();
  // The link gives the title: the note's own heading is not shown again.
  await expect(embed.getByRole('heading')).toHaveCount(0);
  await expect
    .poll(() => box.content(source.id), { timeout: 5000 })
    .toContain(`![${box.prefix} choisie](note:${target.id})`);

  await page.goto(`/notes/${(await box.note(`# ${box.prefix} phrase\n\nvoir ![[${box.prefix} choisie]] ici`)).id}`);
  await expect(page.locator(`.ProseMirror a[data-note="${target.id}"]`)).toBeVisible();
  await expect(page.locator('.note-embed')).toHaveCount(0);
});

test('a [[link]] typed by hand leaves no bracket around it', async ({ page }) => {
  const source = await box.note(`# ${box.prefix} seule\n\nAvant`);
  await page.goto(`/notes/${source.id}`);
  await cursorAtEnd(page.locator('.ProseMirror'));
  await page.keyboard.type(' [[personne]]');
  await page.keyboard.press('Escape');
  await expect.poll(() => box.content(source.id), { timeout: 5000 }).toContain('Avant [[personne]]');
  expect(await box.content(source.id)).not.toContain('\\[');
});

test('![[Note#Heading]] shows that section alone, its tables and callouts drawn (#105)', async ({ page }) => {
  const target = await box.note(
    `# ${box.prefix} source\n\navant\n\n## Lait\n\n| a | b |\n| --- | --- |\n| 1 | 2 |\n\n> [!tip]\n> astuce\n\n## Pain\n\nbaguette`,
  );
  const source = await box.note(`# ${box.prefix} hôte\n\n![[${box.prefix} source#Lait]]`);
  await page.goto(`/notes/${source.id}`);

  const embed = page.locator('.note-embed');
  await expect(embed.locator('table td').getByText('2')).toBeVisible();
  await expect(embed.locator('blockquote[data-callout="tip"]')).toContainText('astuce');
  await expect(embed.getByText('avant')).toHaveCount(0);
  await expect(embed.getByText('baguette')).toHaveCount(0);
  expect(await box.content(source.id)).toContain(`(note:${target.id}#Lait)`);
});
