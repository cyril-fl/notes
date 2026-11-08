import { expect, test } from '@playwright/test';
import { API, APP } from '../../playwright.config';
import { sandbox } from './helpers';

let box: Awaited<ReturnType<typeof sandbox>>;
test.beforeEach(async () => (box = await sandbox()));
test.afterEach(async () => box.dispose());

// The export goes back in as an import: its folders are all taken, so the import is refused and creates nothing.
test('the settings export the notes and import a zip', async ({ page }) => {
  await box.note(`# Exportée ${box.prefix}`);
  await page.goto('/settings');

  const download = page.waitForEvent('download');
  await page.getByRole('link', { name: 'Exporter' }).click();
  const zip = await (await download).path();
  expect(zip).toBeTruthy();

  await page.locator('input[type=file]').setInputFiles(zip);
  await expect(page.getByText('Un dossier du zip existe déjà').first()).toBeVisible();
});

test('a zip is imported into the chosen folder', async ({ page }) => {
  await page.goto('/settings');
  await page.getByRole('button', { name: 'Importer dans' }).click();
  await page.getByRole('option', { name: `/${box.prefix}` }).click();
  await page.locator('input[type=file]').setInputFiles('test/e2e/fixtures/notes.zip');
  await expect(page.getByText('1 note(s) et 1 dossier(s) importés').first()).toBeVisible();

  await page.goto(`/folders/${box.folder.id}`);
  await expect(page.getByRole('main').getByText('Importé')).toBeVisible();
});

// The language is the account's: without the cookie that remembers it, the API gives it back after a reload.
test('the language of the interface is a setting of the account', async ({ page }) => {
  try {
    await page.goto('/settings');
    await page.getByRole('button', { name: 'Langue' }).click();
    await page.getByRole('option', { name: 'English' }).click();
    await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible();

    await page.context().clearCookies({ name: 'notes-locale' });
    await page.reload();
    await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible();
  } finally {
    // The other tests read French.
    await box.api.patch('/api/me', { data: { locale: 'fr' } });
  }
});

// A theme and the trash retention are saved to the account, then set back as they were.
test('the settings save the theme and the trash retention to the account', async ({ page, request }) => {
  const api = { headers: { Origin: APP } };
  const me = async () => (await request.get(`${API}/api/me`, api)).json();
  const before = await me();
  const primary = () =>
    page.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue('--ui-color-primary-500').trim());
  try {
    await page.goto('/settings');
    const notes = await primary();
    await page.getByRole('button', { name: 'Dracula' }).click();
    await expect(page.getByRole('button', { name: 'Dracula' })).toHaveAttribute('aria-pressed', 'true');
    await expect.poll(primary).not.toBe(notes);

    await page.getByRole('button', { name: 'Durée de conservation' }).click();
    await page.getByRole('option', { name: 'Jamais effacés' }).click();

    await expect.poll(me).toMatchObject({ theme: 'dracula', trash_retention_days: 0 });
  } finally {
    await request.patch(`${API}/api/me`, {
      ...api,
      data: { theme: before.theme, trash_retention_days: before.trash_retention_days },
    });
  }
});

// A snippet goes in with ;; and its #tag becomes a tag, then starts a new note from T. Set back to none afterwards.
test('snippets are made in the settings and used in the notes', async ({ page }) => {
  try {
    await page.goto('/settings');
    await page.getByRole('button', { name: 'Ajouter' }).click();
    await page.getByRole('textbox', { name: 'Nom' }).fill(`${box.prefix} réunion`);
    await page.getByRole('textbox', { name: 'Contenu' }).fill('## Réunion #e2esnippet');
    await page.getByRole('button', { name: 'Enregistrer' }).click();
    await expect(page.getByText(`${box.prefix} réunion`)).toBeVisible();

    const note = await box.note(`# ${box.prefix} snippet\n\nTexte`);
    await page.goto(`/notes/${note.id}`);
    const editor = page.locator('.ProseMirror');
    await editor.click();
    await page.keyboard.press('ControlOrMeta+End');
    await page.keyboard.type(' ;;réunion');
    await page.getByRole('option', { name: new RegExp(`${box.prefix} réunion`) }).click();
    await expect.poll(() => box.content(note.id), { timeout: 5000 }).toContain('Réunion #e2esnippet');
    await expect
      .poll(async () => (await (await box.api.get(`/api/notes/${note.id}`)).json()).tags)
      .toContain('e2esnippet');

    await page.goto(`/folders/${box.folder.id}`);
    await expect(page.getByRole('heading', { name: box.prefix })).toBeVisible();
    await page.keyboard.press('t');
    await page.getByRole('option', { name: `${box.prefix} réunion` }).click();
    await expect(page).toHaveURL(/\/notes\/new\?/);
    await expect(editor.getByRole('heading', { name: 'Réunion' })).toBeVisible();
  } finally {
    await box.api.patch('/api/me', { data: { snippets: [] } });
  }
});

// The focus mode and the context panel are Pro features: hidden until the Pro mode is on; the focus mode then hides the sidebar and the panel.
test('the Pro mode unlocks the focus mode and the context panel', async ({ page }) => {
  const note = await box.note(`# Focus ${box.prefix}`);
  await box.note(`# Neighbor ${box.prefix}`);
  try {
    await page.goto(`/notes/${note.id}`);
    await expect(page.getByRole('textbox')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Mode focus' })).toHaveCount(0);
    await expect(page.getByRole('navigation', { name: 'Plan' })).toHaveCount(0);

    await page.goto('/settings');
    await page.getByRole('switch', { name: 'Activer le mode Pro' }).click();
    await expect.poll(async () => (await (await box.api.get('/api/me')).json()).pro).toBe(true);

    await page.goto(`/notes/${note.id}`);
    const sidebar = page.getByRole('link', { name: 'Graphe', exact: true }).first();
    await expect(sidebar).toBeVisible();
    const outline = page.getByRole('navigation', { name: 'Plan' });
    await expect(outline.getByRole('button', { name: `Focus ${box.prefix}` })).toBeVisible();
    // A [[link]] typed in the note draws its neighbor in the panel, without a reload.
    const panel = page.getByRole('complementary', { name: /panneau de contexte/ });
    await expect(panel.getByRole('heading', { name: 'Graphe' })).toHaveCount(0);
    await page.getByRole('textbox').click();
    await page.keyboard.press('End');
    await page.keyboard.type(` [[Neighbor ${box.prefix}]]`);
    await expect(panel.getByRole('heading', { name: 'Graphe' })).toBeVisible({ timeout: 10_000 });
    await page.getByRole('button', { name: 'Mode focus' }).click();
    await expect(sidebar).toBeHidden();
    await expect(outline).toBeHidden();
    await page.getByRole('button', { name: 'Quitter le mode focus' }).click();
    await expect(sidebar).toBeVisible();
  } finally {
    await box.api.patch('/api/me', { data: { pro: false } });
  }
});
