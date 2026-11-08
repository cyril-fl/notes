import { expect, test } from '@playwright/test';
import { cursorAtEnd, sandbox } from './helpers';

let box: Awaited<ReturnType<typeof sandbox>>;
test.beforeEach(async () => (box = await sandbox()));
test.afterEach(async () => box.dispose());

// The Pro context panel of a note: its outline scrolls the text, its linked notes and mini-graph follow a typed
// [[link]] on both ends, ⌘. hides it with the sidebar, and it goes away with the Pro mode.
test('the Pro context panel follows the note', async ({ page }) => {
  const filler = Array.from({ length: 60 }, (_, i) => `Paragraphe ${i}.`).join('\n\n');
  const note = await box.note(`# Panel ${box.prefix}\n\n${filler}\n\n## Bottom ${box.prefix}`);
  const target = await box.note(`# Target ${box.prefix}`);
  const panel = page.getByRole('complementary', { name: /panneau de contexte/ });
  try {
    await box.api.patch('/api/me', { data: { pro: true } });
    await page.goto(`/notes/${note.id}`);

    const outline = panel.getByRole('navigation', { name: 'Plan' });
    await expect(outline.getByRole('button', { name: `Panel ${box.prefix}` })).toBeVisible();
    const bottom = page.getByRole('textbox').getByRole('heading', { name: `Bottom ${box.prefix}` });
    await expect(bottom).not.toBeInViewport();
    await outline.getByRole('button', { name: `Bottom ${box.prefix}` }).click();
    await expect(bottom).toBeInViewport();

    // Linked notes and the mini-graph show up once the [[link]] is saved, without a reload.
    await expect(panel.getByText('Liens', { exact: true })).toHaveCount(0);
    await cursorAtEnd(page.getByRole('textbox'), `Bottom ${box.prefix}`);
    await page.keyboard.type(` [[Target ${box.prefix}]]`);
    await expect(panel.getByText('Liens', { exact: true })).toBeVisible({ timeout: 10_000 });
    const graph = panel.locator('section').filter({ has: page.getByRole('heading', { name: 'Graphe' }) });
    await expect(graph.getByRole('link', { name: `Panel ${box.prefix}`, exact: true })).toBeVisible();
    await expect(graph.getByRole('link', { name: `Target ${box.prefix}`, exact: true })).toBeVisible();

    // The other end shows the backlink.
    await page.goto(`/notes/${target.id}`);
    await expect(panel.getByText('Rétroliens', { exact: true })).toBeVisible();
    await expect(panel.getByRole('link', { name: `Panel ${box.prefix}` }).first()).toBeVisible();

    const sidebar = page.getByRole('link', { name: 'Graphe', exact: true }).first();
    await page.keyboard.press('ControlOrMeta+Period');
    await expect(sidebar).toBeHidden();
    await expect(panel).toBeHidden();
    await page.keyboard.press('ControlOrMeta+Period');
    await expect(panel).toBeVisible();

    await box.api.patch('/api/me', { data: { pro: false } });
    await page.reload();
    await expect(page.getByRole('textbox')).toBeVisible();
    await expect(panel.getByRole('navigation', { name: 'Plan' })).toHaveCount(0);
  } finally {
    await box.api.patch('/api/me', { data: { pro: false } });
  }
});
