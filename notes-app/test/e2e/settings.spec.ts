import { expect, test } from '@playwright/test';
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
