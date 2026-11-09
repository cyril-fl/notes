import { expect, test } from '@playwright/test';
import { cursorAtEnd, sandbox } from './helpers';

let box: Awaited<ReturnType<typeof sandbox>>;
test.beforeEach(async () => (box = await sandbox()));
test.afterEach(async () => box.dispose());

test('[[Note# offers its headings, the link opens the note at that heading (#85)', async ({ page }) => {
  const target = await box.note(`# ${box.prefix} cible\n\n${'texte\n\n'.repeat(80)}## Fin\n\nla fin`);
  const source = await box.note(`# ${box.prefix} source\n\nVoir`);
  await page.goto(`/notes/${source.id}`);
  const editor = page.locator('.ProseMirror');
  await cursorAtEnd(editor);
  // Its headings are read as soon as the note is named, the menu shows them on the next keys, typed as a person would.
  await page.keyboard.type(` [[${box.prefix} cible`);
  await page.keyboard.type('#Fi', { delay: 150 });
  await expect(page.getByRole('option')).toHaveCount(1);
  await page.getByRole('option', { name: new RegExp(`${box.prefix} cible#Fin`) }).click();
  const link = editor.locator(`a[data-note="${target.id}"][data-heading="Fin"]`);
  await expect(link).toBeVisible();

  await expect.poll(() => box.content(source.id), { timeout: 5000 }).toContain(`](note:${target.id}#Fin)`);
  await link.click();
  await expect(page).toHaveURL(`/notes/${target.id}#Fin`);
  await expect(page.getByRole('heading', { name: 'Fin' })).toBeInViewport();
});

test('a [[Note#Heading]] typed by hand waits, then the API points it to the note and keeps the heading', async ({
  page,
}) => {
  const target = await box.note(`# ${box.prefix} main\n\n## Lait`);
  const source = await box.note(`# ${box.prefix} tapé\n\nVoir`);
  await page.goto(`/notes/${source.id}`);
  const editor = page.locator('.ProseMirror');
  await cursorAtEnd(editor);
  await page.keyboard.type(` [[${box.prefix} main#Lait]]`);
  await page.keyboard.press('Escape');
  await expect.poll(() => box.content(source.id), { timeout: 5000 }).toContain(`](note:${target.id}#Lait)`);
  await expect(editor.locator(`a[data-note="${target.id}"][data-heading="Lait"]`)).toBeVisible();
});

test('[[# offers the headings of the note itself, the link scrolls to it without leaving the note (#100)', async ({
  page,
}) => {
  const note = await box.note(`# ${box.prefix} même\n\nVoir\n\n${'texte\n\n'.repeat(80)}## Bas\n\nla fin`);
  await page.goto(`/notes/${note.id}`);
  const editor = page.locator('.ProseMirror');
  await editor.getByText('Voir').click();
  await page.keyboard.press('End');
  await page.keyboard.type(' [[#Ba', { delay: 150 });
  await page.getByRole('option', { name: '#Bas' }).click();

  await expect.poll(() => box.content(note.id), { timeout: 5000 }).toContain('Voir [[#Bas]]');
  await editor.locator('a[data-wikilink="#Bas"]').click();
  await expect(page).toHaveURL(`/notes/${note.id}`);
  await expect(page.getByRole('heading', { name: 'Bas' })).toBeInViewport();
});
