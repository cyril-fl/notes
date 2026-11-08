import { copyFile, mkdir, writeFile } from 'node:fs/promises';
import { join } from 'node:path';
import { expect, test, type Locator, type Page } from '@playwright/test';
import type { Editor } from '@tiptap/core';
import { API, APP } from '../../playwright.config';
import { sandbox } from './helpers';

let box: Awaited<ReturnType<typeof sandbox>>;
test.beforeEach(async () => (box = await sandbox()));
test.afterEach(async () => box.dispose());

// The import is a modal: a destination, then the files dropped or picked.
async function importFiles(page: Page, files: Parameters<Locator['setInputFiles']>[0], into?: string) {
  await page.getByRole('button', { name: 'Importer', exact: true }).click();
  const dialog = page.getByRole('dialog');
  if (into) {
    await dialog.getByRole('button', { name: 'Importer dans' }).click();
    await page.getByRole('option', { name: into }).click();
  }
  await dialog.locator('input[type=file]:not([webkitdirectory])').setInputFiles(files);
  await dialog.getByRole('button', { name: 'Importer', exact: true }).click();
}

// The export goes back in as an import: its folders are all taken, so the import is refused and creates nothing.
test('the settings export the notes and import a zip', async ({ page }) => {
  await box.note(`# Exportée ${box.prefix}`);
  await page.goto('/settings');

  const download = page.waitForEvent('download');
  await page.getByRole('link', { name: 'Exporter' }).click();
  // Saved under its .zip name: the dialog sends a lone zip as it is, anything else as loose files.
  const zip = test.info().outputPath('notes.zip');
  await (await download).saveAs(zip);

  await importFiles(page, zip);
  await expect(page.getByText('Un dossier du zip existe déjà').first()).toBeVisible();
});

test('a zip is imported into the chosen folder', async ({ page }) => {
  await page.goto('/settings');
  await importFiles(page, 'test/e2e/fixtures/notes.zip', `/${box.prefix}`);
  await expect(page.getByText('1 note(s) et 1 dossier(s) importés').first()).toBeVisible();
  await expect(page.getByRole('dialog')).toBeHidden();

  await page.goto(`/folders/${box.folder.id}`);
  await expect(page.getByRole('main').getByText('Importé')).toBeVisible();
});

// Loose files go as a form; the paths of a picked folder are the API's test (Playwright sets no webkitRelativePath).
test('loose .md files are imported into the chosen folder', async ({ page }) => {
  await page.goto('/settings');
  const md = (name: string) => ({ name, mimeType: 'text/markdown', buffer: Buffer.from(`# ${name}`) });
  await importFiles(page, [md('Une.md'), md('Deux.md')], `/${box.prefix}`);
  await expect(page.getByText('2 note(s) et 0 dossier(s) importés').first()).toBeVisible();

  await page.goto(`/folders/${box.folder.id}`);
  await expect(page.getByRole('main').getByText('Une.md')).toBeVisible();
  await expect(page.getByRole('main').getByText('Deux.md')).toBeVisible();
});

// A picked folder keeps its notes but not its zips, which would gray out the button.
test('a picked folder holding a zip imports its notes', async ({ page }) => {
  const vault = test.info().outputPath('Coffre');
  await mkdir(join(vault, 'attachments'), { recursive: true });
  await writeFile(join(vault, 'Une.md'), '# Une');
  await copyFile('test/e2e/fixtures/notes.zip', join(vault, 'attachments', 'backup.zip'));

  await page.goto('/settings');
  await page.getByRole('button', { name: 'Importer', exact: true }).click();
  const dialog = page.getByRole('dialog');
  await dialog.getByRole('button', { name: 'Importer dans' }).click();
  await page.getByRole('option', { name: `/${box.prefix}` }).click();
  await dialog.locator('input[webkitdirectory]').setInputFiles(vault);
  await expect(dialog.getByText('Une.md')).toBeVisible();
  await expect(dialog.getByText('backup.zip')).toBeHidden();
  await dialog.getByRole('button', { name: 'Importer', exact: true }).click();
  await expect(page.getByText('1 note(s) et 1 dossier(s) importés').first()).toBeVisible();
});

// Past the API's 200 MB, the toast gives the limit in French, not the API's message (#58).
test('an import too large says its limit', async ({ page }) => {
  const zip = test.info().outputPath('gros.zip');
  await writeFile(zip, Buffer.alloc(210 << 20));
  await page.goto('/settings');
  await importFiles(page, zip);
  await expect(page.getByText('Import trop gros : 200 Mo au plus').first()).toBeVisible();
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
    // The sidebar becomes the settings nav, one page per section.
    await page.goto('/settings');
    await expect(page).toHaveURL(/\/settings\/general$/);
    await page.getByRole('link', { name: 'Apparence' }).click();
    await expect(page).toHaveURL(/\/settings\/appearance$/);
    await expect(page.getByRole('heading', { name: 'Apparence' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Langue' })).toHaveCount(0);
    await page.getByRole('link', { name: 'Retour aux notes' }).click();
    await expect(page.getByRole('link', { name: 'Graphe', exact: true }).first()).toBeVisible();

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
    // A renamed note shows its new title in the panel's graph, without a reload.
    await page.getByRole('textbox').getByRole('heading').click();
    await page.keyboard.press('Home');
    await page.keyboard.type('Renamed ');
    await expect(panel.getByRole('link', { name: `Renamed Focus ${box.prefix}` })).toBeVisible({ timeout: 10_000 });
    await page.getByRole('button', { name: 'Mode focus' }).click();
    await expect(sidebar).toBeHidden();
    await expect(outline).toBeHidden();
    await page.getByRole('button', { name: 'Quitter le mode focus' }).click();
    await expect(sidebar).toBeVisible();
  } finally {
    await box.api.patch('/api/me', { data: { pro: false } });
  }
});

// The / menu is the Pro mode's: without it a / stays text, with it the line becomes the block picked in a submenu.
test('the Pro mode adds the / commands to the editor', async ({ page }) => {
  const note = await box.note(`# Slash ${box.prefix}\n\nTexte`);
  const editor = page.locator('.ProseMirror');
  try {
    await page.goto(`/notes/${note.id}`);
    await editor.click();
    await page.keyboard.press('ControlOrMeta+End');
    await page.keyboard.type(' /');
    await expect(page.getByRole('menuitem', { name: 'Bloc de code' })).toHaveCount(0);

    await box.api.patch('/api/me', { data: { pro: true } });
    await page.reload();
    await editor.click();
    await page.keyboard.press('ControlOrMeta+End');
    await page.keyboard.press('Enter');
    await page.keyboard.type('/');
    // The menu has the focus: a letter is for it, not a shortcut of the page (t opens the snippets).
    await expect(page.getByRole('menuitem', { name: 'Titres' })).toBeVisible();
    await page.keyboard.type('t');
    await expect(page.getByRole('dialog')).toHaveCount(0);
    // A backspace takes back the / and closes the menu.
    await page.keyboard.press('Backspace');
    await expect(page.getByRole('menuitem', { name: 'Titres' })).toHaveCount(0);
    await expect(editor).not.toContainText('/');
    // A second / is written, the menu closed.
    await page.keyboard.type('//');
    await expect(page.getByRole('menuitem', { name: 'Titres' })).toHaveCount(0);
    await expect(editor).toContainText('//');
    await page.keyboard.press('Backspace');
    await page.keyboard.press('Backspace');
    await page.keyboard.type('/');
    await page.getByRole('menuitem', { name: 'Titres' }).hover();
    await page.getByRole('menuitem', { name: 'Titre 2' }).click();
    await page.keyboard.type('Sous-titre');
    await expect.poll(() => box.content(note.id), { timeout: 5000 }).toContain('## Sous-titre');
  } finally {
    await box.api.patch('/api/me', { data: { pro: false } });
  }
});

// A table comes from the / menu, grows from the bar over it, and is saved as a markdown table.
test('the Pro mode adds the tables to the editor', async ({ page }) => {
  const note = await box.note(`# Tableau ${box.prefix}\n\nTexte`);
  const editor = page.locator('.ProseMirror');
  try {
    await box.api.patch('/api/me', { data: { pro: true } });
    await page.goto(`/notes/${note.id}`);
    await editor.click();
    await page.keyboard.press('ControlOrMeta+End');
    await page.keyboard.press('Enter');
    await page.keyboard.type('/');
    await page.getByRole('menuitem', { name: 'Tableau' }).click();
    await page.keyboard.type('Nom');
    await page.keyboard.press('Tab');
    await page.keyboard.type('Âge');
    await page.getByRole('button', { name: 'Ajouter une ligne' }).click();
    await expect(editor.locator('tr')).toHaveCount(4);
    await expect.poll(() => box.content(note.id), { timeout: 5000 }).toMatch(/\| *Nom *\| *Âge *\|/);

    await page.reload();
    await expect(editor.getByRole('columnheader', { name: 'Âge' })).toBeVisible();
  } finally {
    await box.api.patch('/api/me', { data: { pro: false } });
  }
});

// A code block is colored by its language in the Pro mode only, and its markdown stays the same.
test('the Pro mode colors the code blocks', async ({ page }) => {
  const note = await box.note(`# Code ${box.prefix}\n\n\`\`\`go\nfunc main() {}\n\`\`\``);
  const keyword = page.locator('.ProseMirror pre .hljs-keyword');
  try {
    await page.goto(`/notes/${note.id}`);
    await expect(page.locator('.ProseMirror pre')).toContainText('func main');
    await expect(keyword).toHaveCount(0);

    await box.api.patch('/api/me', { data: { pro: true } });
    await page.reload();
    await expect(keyword.first()).toHaveText('func');
    await page.locator('.ProseMirror pre').click();
    await page.keyboard.press('End');
    await page.keyboard.type(' // ok');
    await expect.poll(() => box.content(note.id), { timeout: 5000 }).toContain('```go\nfunc main() {} // ok\n```');
  } finally {
    await box.api.patch('/api/me', { data: { pro: false } });
  }
});

// The word count is the Pro mode's, and follows the typing.
test('the Pro mode counts the words of the note', async ({ page }) => {
  const note = await box.note(`# Mots ${box.prefix}\n\nUn deux trois`);
  const count = page.getByText(/^\d+ mots$/);
  try {
    await page.goto(`/notes/${note.id}`);
    await expect(page.locator('.ProseMirror')).toContainText('Un deux trois');
    await expect(count).toHaveCount(0);

    await box.api.patch('/api/me', { data: { pro: true } });
    await page.reload();
    const before = parseInt((await count.textContent()) ?? '');
    await page.locator('.ProseMirror').getByText('Un deux trois').click();
    await page.keyboard.press('End');
    await page.keyboard.type(' quatre, cinq !');
    await expect(count).toHaveText(`${before + 2} mots`);
  } finally {
    await box.api.patch('/api/me', { data: { pro: false } });
  }
});

// A callout comes from the / menu and is saved as an Obsidian and GitHub callout; one already in a note keeps its title.
test('the Pro mode adds the callouts to the editor', async ({ page }) => {
  const note = await box.note(`# Encadré ${box.prefix}\n\n> [!tip] Astuce\n> Lire la suite\n\nTexte`);
  const editor = page.locator('.ProseMirror');
  try {
    await box.api.patch('/api/me', { data: { pro: true } });
    await page.goto(`/notes/${note.id}`);
    await expect(editor.locator('blockquote[data-callout=tip]')).toContainText('Lire la suite');
    await editor.getByText('Texte', { exact: true }).click();
    // ProseMirror reads the click's selection a moment later: typed sooner, Enter splits the title.
    await expect
      .poll(() =>
        editor.evaluate(
          (e) => (e as { editor?: Editor } & HTMLElement).editor?.state.selection.$from.parent.textContent,
        ),
      )
      .toBe('Texte');
    await page.keyboard.press('End');
    await page.keyboard.press('Enter');
    await page.keyboard.type('/');
    await page.getByRole('menuitem', { name: 'Encadrés' }).hover();
    await page.getByRole('menuitem', { name: 'Attention' }).click();
    await page.keyboard.type('Prudence');
    await expect.poll(() => box.content(note.id), { timeout: 5000 }).toContain('> [!warning]\n> Prudence');
    expect(await box.content(note.id)).toContain('> [!tip] Astuce\n> Lire la suite');
  } finally {
    await box.api.patch('/api/me', { data: { pro: false } });
  }
});

// A right click opens the / commands, the callouts in a submenu, applied to the line clicked.
test('the Pro mode opens the / commands on a right click', async ({ page }) => {
  const note = await box.note(`# Clic droit ${box.prefix}\n\nPrudence`);
  const line = page.locator('.ProseMirror').getByText('Prudence', { exact: true });
  try {
    await page.goto(`/notes/${note.id}`);
    await line.click({ button: 'right' });
    await expect(page.getByRole('menu')).toHaveCount(0);

    await box.api.patch('/api/me', { data: { pro: true } });
    await page.reload();
    await line.click({ button: 'right' });
    await page.getByRole('menuitem', { name: 'Encadrés' }).hover();
    await page.getByRole('menuitem', { name: 'Attention' }).click();
    await expect.poll(() => box.content(note.id), { timeout: 5000 }).toContain('> [!warning]\n> Prudence');
  } finally {
    await box.api.patch('/api/me', { data: { pro: false } });
  }
});

// A page per section is the Pro mode's: without it a section's URL shows the whole page, and the switch moves between both.
test('the settings follow the Pro mode, one page or a page per section', async ({ page }) => {
  try {
    await box.api.patch('/api/me', { data: { pro: false } });
    await page.goto('/settings/pro');
    await expect(page).toHaveURL(/\/settings$/);

    await page.getByRole('switch', { name: 'Activer le mode Pro' }).click();
    await expect(page).toHaveURL(/\/settings\/pro$/);
    await expect(page.getByRole('link', { name: 'Retour aux notes' })).toBeVisible();
    await page.goto('/settings/nope');
    await expect(page).toHaveURL(/\/settings\/general$/);

    await page.goto('/settings/pro');
    await page.getByRole('switch', { name: 'Activer le mode Pro' }).click();
    await expect(page).toHaveURL(/\/settings$/);
    await expect(page.getByRole('heading', { name: 'Réglages' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Retour aux notes' })).toHaveCount(0);
  } finally {
    await box.api.patch('/api/me', { data: { pro: false } });
  }
});
