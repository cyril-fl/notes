import { expect, test } from '@playwright/test';
import { sandbox } from './helpers';

let box: Awaited<ReturnType<typeof sandbox>>;
test.beforeEach(async () => (box = await sandbox()));
test.afterEach(async () => box.dispose());

// Zoom spreads the notes from the center: dezoomed, two linked notes come closer, and Recentrer puts them back.
// Dezoomed, not zoomed: the server lays the notes out to the edges, zoomed one leaves the view and loses its link.
// The neighborhood of a note, not every note: on a full account the titles thin out and a note is a bare dot, without link.
test('the graph zooms and recenters', async ({ page }) => {
  const tag = box.prefix.replace('-', '');
  const note = await box.note(`# Un ${box.prefix}\n\n#${tag}`);
  await box.note(`# Deux ${box.prefix}\n\n#${tag}`);
  await page.goto(`/graph?note=${note.id}`);
  await expect(page.locator('[data-settled]')).toHaveCount(1); // The layout moves the notes until then.
  const one = page.getByRole('main').getByRole('link', { name: `Un ${box.prefix}` });
  const two = page.getByRole('main').getByRole('link', { name: `Deux ${box.prefix}` });
  const gap = async () => {
    const [a, b] = [await one.boundingBox(), await two.boundingBox()];
    return Math.hypot(a!.x - b!.x, a!.y - b!.y);
  };

  const before = await gap();
  await page.getByRole('button', { name: 'Dézoomer' }).click();
  expect(await gap()).toBeLessThan(before / 1.3);
  await page.getByRole('button', { name: 'Recentrer' }).click();
  expect(Math.abs((await gap()) - before)).toBeLessThan(1);
  await box.api.delete(`/api/tags/${tag}`); // Deleting the notes leaves their tag.
});

// From a note, the graph is its neighborhood, as the mini graph; Tout le graphe shows every note.
test('the graph of a note shows its neighbors, then every note', async ({ page }) => {
  const tag = box.prefix.replace('-', '');
  const one = await box.note(`# Un ${box.prefix}\n\n#${tag}`);
  await box.note(`# Deux ${box.prefix}\n\n#${tag}`);
  await box.note(`# Seule ${box.prefix}`);
  const main = page.getByRole('main');

  await page.goto(`/graph?note=${one.id}`);
  await expect(page.locator('[data-settled]')).toHaveCount(1);
  await expect(main.getByRole('link', { name: `Deux ${box.prefix}` })).toBeVisible();
  await expect(main.getByRole('link', { name: `Seule ${box.prefix}` })).toHaveCount(0);

  await main.getByRole('link', { name: 'Tout le graphe' }).click();
  await expect(page.locator('[data-settled]')).toHaveCount(1);
  await expect(main.getByRole('link', { name: `Seule ${box.prefix}` })).toBeVisible();
  await box.api.delete(`/api/tags/${tag}`);
});

// The server keeps the layout (#153): another device, with nothing in its browser, gets the same map.
test('another device gets the same layout of the graph', async ({ page, browser }) => {
  const tag = box.prefix.replace('-', '');
  await box.note(`# Un ${box.prefix}\n\n#${tag}`);
  await box.note(`# Deux ${box.prefix}\n\n#${tag}`);
  const place = async (on: typeof page) => {
    await on.goto('/graph');
    await expect(on.locator('[data-settled]')).toHaveCount(1);
    return on
      .getByRole('main')
      .getByRole('link', { name: `Un ${box.prefix}` })
      .boundingBox();
  };

  const here = await place(page);
  const device = await browser.newContext({ storageState: 'test-results/.session.json', locale: 'fr-FR' });
  try {
    expect(await place(await device.newPage())).toEqual(here);
  } finally {
    await device.close();
  }
  await box.api.delete(`/api/tags/${tag}`);
});
