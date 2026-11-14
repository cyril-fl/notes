// Times the API and the pages of the account of generated notes (make perf, #138), and prints a table to paste in an issue.
// node scripts/perf.mjs  (PERF_NAME / PERF_PASSWORD: perf / perf-password by default)
// Needs `make up` and the account, made by `notes user perf` (make perf does both).
import { chromium, request } from '@playwright/test';

const APP = process.env.E2E_APP ?? 'http://localhost:3000';
const API = process.env.E2E_API ?? 'http://localhost:8080';
const NAME = process.env.PERF_NAME ?? 'perf';
const PASSWORD = process.env.PERF_PASSWORD ?? 'perf-password';
// A page slower than this is written as such: the point is to see it, not to wait for it.
const TIMEOUT = 300_000;

const api = await request.newContext({ baseURL: API, extraHTTPHeaders: { Origin: APP }, timeout: TIMEOUT });
const login = await api.post('/api/login', { data: { name: NAME, password: PASSWORD } });
if (!login.ok()) throw new Error(`login ${NAME}: ${login.status()}, run make perf to create the account`);
const storageState = await api.storageState();

const get = async (path) => (await api.get(path)).json();
const folders = await get('/api/folders');
const big = folders.find((folder) => folder.path === '/Archives') ?? folders[0];
const [tag] = await get('/api/graph');
const { notes: tagged } = await get(`/api/tags/${encodeURIComponent(tag.value)}`);
const note = tagged[0];
const word = 'projet';
const notes = (await get('/api/notes')).length;

const rows = [];
const format = (ms) =>
  ms === undefined ? `> ${TIMEOUT / 1000} s` : ms < 1000 ? `${ms} ms` : `${(ms / 1000).toFixed(1)} s`;
const size = (bytes) =>
  bytes < 1024
    ? `${bytes} o`
    : bytes < 1 << 20
      ? `${(bytes / 1024).toFixed(0)} Ko`
      : `${(bytes / (1 << 20)).toFixed(1)} Mo`;

for (const path of [
  `/api/folders/${big.id}`,
  '/api/items',
  '/api/folders',
  '/api/graph',
  '/api/graph/notes',
  '/api/notes',
  `/api/notes/${note.id}/neighbors`,
  `/api/search?q=${word}`,
  `/api/tags/${encodeURIComponent(tag.value)}`,
  '/api/recent',
]) {
  const start = Date.now();
  try {
    const response = await api.get(path);
    const body = await response.body();
    rows.push([
      `\`GET ${path.replace(big.id, '{archives}').replace(note.id, '{note}')}\``,
      format(Date.now() - start),
      `${size(body.length)}, ${response.status()}`,
    ]);
  } catch {
    rows.push([`\`GET ${path}\``, format(undefined), '']);
  }
}

const browser = await chromium.launch({ channel: 'chrome' });
// A fresh context per page: nothing cached from the page before (the graph's saved layout aside, measured on purpose).
// linger: ms left to the page after it is ready, for the graph to save its layout before the page closes.
async function visit(label, path, ready, context, linger = 0) {
  const own = !context;
  context ??= await browser.newContext({ storageState, locale: 'fr-FR', viewport: { width: 1280, height: 800 } });
  const page = await context.newPage();
  const start = Date.now();
  try {
    await page.goto(APP + path, { timeout: TIMEOUT });
    await (ready
      ? page.waitForSelector(ready, { timeout: TIMEOUT })
      : page.waitForLoadState('networkidle', { timeout: TIMEOUT }));
    const ms = Date.now() - start;
    rows.push([label, format(ms), `${await page.evaluate(() => document.getElementsByTagName('*').length)} nœuds DOM`]);
  } catch {
    rows.push([label, format(undefined), '']);
  }
  await page.waitForTimeout(linger);
  await page.close();
  if (own) await context.close();
}

await visit(`Le dossier Archives (${big.path})`, `/folders/${big.id}`);
await visit(`Une note du tag #${tag.value} (${tag.notes} notes)`, `/notes/${note.id}`);
await visit(`La page du tag #${tag.value}`, `/tags/${encodeURIComponent(tag.value)}`);
await visit(`La recherche « ${word} »`, `/search?q=${word}`);
const graph = await browser.newContext({ storageState, locale: 'fr-FR', viewport: { width: 1280, height: 800 } });
await visit('Le graphe, première visite', '/graph', '[data-settled]', graph, 10_000);
await visit('Le graphe, revisite (mise en page gardée)', '/graph', '[data-settled]', graph);
await graph.close();
await browser.close();

console.log(`Compte \`${NAME}\` : ${notes} notes, ${new Date().toISOString().slice(0, 10)}.\n`);
console.log('| Mesure | Temps | Taille ou DOM |\n|---|---|---|');
for (const row of rows) console.log(`| ${row.join(' | ')} |`);
