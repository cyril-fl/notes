import { expect, test } from '@playwright/test';
import { API, APP } from '../../playwright.config';
import { sandbox, throwaway } from './helpers';

// Each test acts on an account of its own: the shared e2e account and its session stay untouched.
let account: Awaited<ReturnType<typeof throwaway>>;
test.beforeEach(async () => (account = await throwaway()));
test.afterEach(async () => account.dispose());

test('deleting all the data asks the name of the account, then empties it and keeps it logged in', async ({
  browser,
}) => {
  const folder = await (await account.api.post('/api/folders', { data: { title: 'Doomed' } })).json();
  await account.api.post('/api/notes', { data: { content: '# Doomed note', parent_id: folder.id } });
  await account.api.post('/api/tags', { data: { value: 'doomed' } });

  const page = await (await browser.newContext({ storageState: account.storageState })).newPage();
  await page.goto('/settings');
  await page.getByRole('button', { name: 'Supprimer', exact: true }).click();
  const dialog = page.getByRole('dialog');
  const action = dialog.getByRole('button', { name: 'Supprimer toutes mes données' });
  await expect(action).toBeDisabled();
  await dialog.getByRole('textbox').fill(`${account.name}x`);
  await expect(action).toBeDisabled();
  await dialog.getByRole('textbox').fill(account.name);
  await action.click();
  await expect(page.getByText('Toutes les données ont été supprimées').first()).toBeVisible();

  expect(await (await account.api.get('/api/items')).json()).toEqual([]);
  expect(await (await account.api.get('/api/tags')).json()).toEqual([]);
  expect(await (await account.api.get('/api/trash')).json()).toEqual([]);
  expect((await account.api.get('/api/me')).ok()).toBe(true);
  await page.context().close();
});

test('a new password logs the other devices out, and only it logs in', async ({ browser }) => {
  const page = await (await browser.newContext({ storageState: account.storageState })).newPage();
  const login = (password: string) => account.api.post('/api/login', { data: { name: account.name, password } });
  // A second session for account.api: another device than the page.
  await login(account.password);
  await page.goto('/account');
  await page.getByLabel('Mot de passe actuel').fill(account.password);
  await page.getByLabel('Nouveau mot de passe').fill('a brand new password');
  await page.getByRole('button', { name: 'Changer le mot de passe' }).click();
  await expect(page.getByText('Mot de passe changé').first()).toBeVisible();

  // The page changing it stays logged in, the other device is logged out.
  await page.reload();
  await expect(page).toHaveURL(/\/account$/);
  expect((await account.api.get('/api/me')).status()).toBe(401);

  expect((await login(account.password)).status()).toBe(401);
  expect((await login('a brand new password')).ok()).toBe(true);
  await page.context().close();
});

test("an account never sees nor touches another account's notes", async ({ browser }) => {
  const box = await sandbox();
  try {
    const note = await box.note(`# ${box.prefix} secret`);

    expect((await account.api.get(`/api/notes/${note.id}`)).status()).toBe(404);
    expect((await account.api.get(`/api/folders/${box.folder.id}`)).status()).toBe(404);
    await account.api.patch(`/api/notes/${note.id}`, { data: { content: '# stolen' } });
    await account.api.delete(`/api/notes/${note.id}`);
    expect(await box.content(note.id)).toBe(`# ${box.prefix} secret`);
    expect(await (await account.api.get('/api/search', { params: { q: box.prefix } })).json()).toEqual([]);

    const page = await (await browser.newContext({ storageState: account.storageState })).newPage();
    await page.goto('/');
    await page.keyboard.press('ControlOrMeta+k');
    await page.keyboard.type(box.prefix);
    await expect(page.getByText(`${box.prefix} secret`)).toHaveCount(0);
    await page.context().close();
  } finally {
    await box.dispose();
  }
});
