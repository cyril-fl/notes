import { expect, test } from '@playwright/test';
import { sandbox } from './helpers';

let box: Awaited<ReturnType<typeof sandbox>>;
test.beforeEach(async () => (box = await sandbox()));
test.afterEach(async () => box.dispose());

// The Pro tabs of the editor: a note opened takes one, a note trashed loses its own, a tab closed opens its
// neighbour, they survive a reload and they go away with the Pro mode.
test('the Pro tabs keep the open notes', async ({ page }) => {
  const first = await box.note(`# First ${box.prefix}`);
  const second = await box.note(`# Second ${box.prefix}`);
  const third = await box.note(`# Third ${box.prefix}`);
  const bar = page.getByRole('navigation', { name: 'Notes ouvertes' });
  try {
    await box.api.patch('/api/me', { data: { plan: 'pro' } });

    // One tab more per note opened: the tab is written once the editor is there, so each load is waited for.
    for (const [index, note] of [first, second, third].entries()) {
      await page.goto(`/notes/${note.id}`);
      await expect(bar.getByRole('link')).toHaveCount(index + 1);
    }

    // Back to the first note by its tab, the three still open.
    await bar.getByRole('link', { name: `First ${box.prefix}` }).click();
    await expect(page.getByRole('textbox').getByRole('heading', { name: `First ${box.prefix}` })).toBeVisible();
    await expect(bar.getByRole('link')).toHaveCount(3);

    // A note trashed from elsewhere leaves the graph, and its tab with it.
    await box.api.delete(`/api/notes/${second.id}`);
    await page.goto(`/notes/${third.id}`);
    await expect(bar.getByRole('link', { name: `Second ${box.prefix}` })).toHaveCount(0);
    await expect(bar.getByRole('link')).toHaveCount(2);

    // The open note's tab closed opens its neighbour, the one before it.
    await bar.getByRole('button', { name: `Fermer l'onglet Third ${box.prefix}` }).click();
    await expect(page).toHaveURL(`/notes/${first.id}`);

    // What is left survives a reload, and the Pro mode takes the bar away.
    await page.reload();
    await expect(bar.getByRole('link', { name: `First ${box.prefix}` })).toBeVisible();
    await expect(bar.getByRole('link')).toHaveCount(1);

    await box.api.patch('/api/me', { data: { plan: 'base' } });
    await page.reload();
    await expect(page.getByRole('textbox')).toBeVisible();
    await expect(bar).toHaveCount(0);
  } finally {
    await box.api.patch('/api/me', { data: { plan: 'base' } });
    await box.api.delete(`/api/trash/${second.id}`);
  }
});
