import { request } from '@playwright/test';
import { API, APP, SESSION } from '../../playwright.config';

/**
 * Logs the e2e account in once, before the tests. Make it after yours (the first account takes the notes made
 * before accounts): notes user add e2e, password E2E_PASSWORD (e2e-password by default).
 */
export default async function login() {
  const api = await request.newContext({ baseURL: API, extraHTTPHeaders: { Origin: APP } });
  const response = await api.post('/api/login', {
    data: { name: process.env.E2E_NAME ?? 'e2e', password: process.env.E2E_PASSWORD ?? 'e2e-password' },
  });
  if (!response.ok())
    throw new Error(
      `e2e login: ${response.status()} ${await response.text()}: create the account with notes user add e2e`,
    );
  await api.storageState({ path: SESSION });
  await api.dispose();
}
