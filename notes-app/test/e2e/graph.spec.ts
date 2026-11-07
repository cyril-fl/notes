import { expect, test } from '@playwright/test';
import { sandbox } from './helpers';

let box: Awaited<ReturnType<typeof sandbox>>;
test.beforeEach(async () => (box = await sandbox()));
test.afterEach(async () => box.dispose());

// Zoom spreads the notes: two linked notes move apart, and Recentrer puts them back.
test('the graph zooms and recenters', async ({ page }) => {
  const tag = box.prefix.replace('-', '');
  await box.note(`# Un ${box.prefix}\n\n#${tag}`);
  await box.note(`# Deux ${box.prefix}\n\n#${tag}`);
  await page.goto('/graph');
  const one = page.getByRole('main').getByRole('link', { name: `Un ${box.prefix}` });
  const two = page.getByRole('main').getByRole('link', { name: `Deux ${box.prefix}` });
  const gap = async () => {
    const [a, b] = [await one.boundingBox(), await two.boundingBox()];
    return Math.hypot(a!.x - b!.x, a!.y - b!.y);
  };

  const before = await gap();
  await page.getByRole('button', { name: 'Zoomer', exact: true }).click();
  expect(await gap()).toBeGreaterThan(before * 1.3);
  await page.getByRole('button', { name: 'Recentrer' }).click();
  expect(Math.abs((await gap()) - before)).toBeLessThan(1);
  await box.api.delete(`/api/tags/${tag}`); // Deleting the notes leaves their tag.
});
