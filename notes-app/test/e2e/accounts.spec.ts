import { execFileSync } from 'node:child_process';
import { expect, test } from '@playwright/test';
import { throwaway } from './helpers';

// An admin of its own: the shared e2e account manages nothing, and this one is deleted with the test.
let account: Awaited<ReturnType<typeof throwaway>>;
test.beforeEach(async () => {
  account = await throwaway();
  sql(`UPDATE users SET admin = true, plan = 'pro' WHERE name = '${account.name}'`);
});
test.afterEach(async () => account.dispose());

/** Runs a statement in the base of the stack, as the throwaway accounts are made and deleted. */
function sql(query: string) {
  execFileSync(
    'docker',
    ['exec', '-i', 'note-postgres', 'sh', '-c', 'psql -q -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" "$POSTGRES_DB"'],
    { input: query },
  );
}

test('the accounts section invites, and the invitation link creates an account the admin can delete', async ({
  browser,
}) => {
  const invited = `e2e-invited-${Date.now()}`;
  const page = await (await browser.newContext({ storageState: account.storageState })).newPage();
  try {
    await page.goto('/settings/accounts');
    await expect(page.getByRole('cell', { name: account.name })).toBeVisible();

    // Every pending invitation shows its link, the newest one first: the list may already hold others.
    await page.getByRole('button', { name: 'Inviter' }).click();
    const link = page.getByRole('textbox', { name: /Lien d.invitation/ }).first();
    await expect(link).toBeVisible();
    const url = await link.inputValue();
    expect(url).toContain('/signup?invite=');

    // Someone else opens it, in a browser without a session, and chooses their name and password.
    const guest = await (await browser.newContext({ storageState: { cookies: [], origins: [] } })).newPage();
    await guest.goto(url);
    await guest.getByRole('textbox', { name: 'Nom' }).fill(invited);
    await guest.getByLabel('Mot de passe', { exact: true }).fill('a-good-password');
    await guest.getByLabel('Mot de passe (à nouveau)').fill('a-good-password');
    await guest.getByRole('button', { name: 'Créer le compte' }).click();
    await expect(guest).toHaveURL(/\/$/);

    // The same link serves nobody else.
    await guest.goto(url);
    await expect(guest).toHaveURL(/\/login/);
    await guest.context().close();

    // The admin sees the account, and deletes it with everything it holds.
    await page.reload();
    await expect(page.getByRole('cell', { name: invited })).toBeVisible();
    await page.getByRole('row', { name: invited }).getByRole('button', { name: 'Supprimer' }).click();
    const dialog = page.getByRole('dialog');
    await dialog.getByRole('textbox').fill(invited);
    await dialog.getByRole('button', { name: 'Supprimer' }).click();
    await expect(page.getByRole('cell', { name: invited })).toBeHidden();
  } finally {
    sql(`DELETE FROM users WHERE name = '${invited}'`);
    await page.context().close();
  }
});

test('without an invitation the sign-up page goes back to the login', async ({ browser }) => {
  const guest = await (await browser.newContext({ storageState: { cookies: [], origins: [] } })).newPage();
  await guest.goto('/signup');
  await expect(guest).toHaveURL(/\/login/);
  // The server invites: the login page offers no sign-up of its own.
  await expect(guest.getByRole('link', { name: 'Créer un compte' })).toBeHidden();
  await guest.context().close();
});
