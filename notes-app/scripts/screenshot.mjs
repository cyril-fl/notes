// Screenshots a page of the running app as the e2e account, its settings changed for the shot, then set back.
// node scripts/screenshot.mjs <out.png> [page path] [settings JSON for PATCH /api/me]
// e.g. node scripts/screenshot.mjs /tmp/nord-dark.png /settings '{"theme":"nord","color_mode":"dark"}'
// Needs `make up` and the session written by a `pnpm e2e` run (test-results/.session.json).
import { chromium, request } from '@playwright/test';

const APP = process.env.E2E_APP ?? 'http://localhost:3000';
const API = process.env.E2E_API ?? 'http://localhost:8080';
const SESSION = 'test-results/.session.json';

const [out, path = '/', settings = '{}'] = process.argv.slice(2);
if (!out) {
  console.error('usage: node scripts/screenshot.mjs <out.png> [page path] [settings JSON]');
  process.exit(2);
}
const changes = JSON.parse(settings);

const api = await request.newContext({ baseURL: API, storageState: SESSION, extraHTTPHeaders: { Origin: APP } });
const me = await api.get('/api/me');
if (!me.ok()) throw new Error(`GET /api/me: ${me.status()}, run pnpm e2e once to log in`);
const before = await me.json();
const browser = await chromium.launch({ channel: 'chrome' });
try {
  if (Object.keys(changes).length) {
    const edit = await api.patch('/api/me', { data: changes });
    if (!edit.ok()) throw new Error(`PATCH /api/me: ${edit.status()} ${await edit.text()}`);
  }
  const page = await browser.newPage({
    storageState: SESSION,
    locale: 'fr-FR',
    viewport: { width: 1100, height: 800 },
  });
  await page.goto(APP + path);
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(500); // the theme and mode of the account apply once /api/me answered
  await page.screenshot({ path: out });
  console.log(out);
} finally {
  // Only the keys changed go back: the account is the one every e2e run uses.
  if (Object.keys(changes).length) {
    await api.patch('/api/me', { data: Object.fromEntries(Object.keys(changes).map((key) => [key, before[key]])) });
  }
  await browser.close();
}
