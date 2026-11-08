import assert from 'node:assert/strict';
import { test } from 'node:test';
import type { Item } from '../app/types/items.ts';
import { useItems } from '../app/composables/useItems.ts';

type Call = { url: string; method?: string; body?: unknown };

// The Nuxt auto-imports useItems reads, faked: an API that records its calls, dialogs that answer as told.
function fake({ path = '/', pinned = [] as string[], answer = undefined as unknown, fails = false } = {}) {
  const calls: Call[] = [];
  const navigations: unknown[] = [];
  let refreshes = 0;
  Object.assign(globalThis, {
    useApi:
      () =>
      async (url: string, options: { method?: string; body?: unknown } = {}) => {
        calls.push({ url, method: options.method, body: options.body });
        if (fails) throw new Error('refused');
        return 'ok';
      },
    useRoute: () => ({ path }),
    useDialogs: () => ({ prompt: async () => answer, confirm: async () => answer }),
    useIcons: () => ({ icons: new Proxy({}, { get: (_, name) => name }) }),
    useI18n: () => ({ t: (key: string) => key }),
    useNuxtData: () => ({ data: { value: pinned.map((id) => ({ id })) } }),
    refreshNuxtData: async () => void refreshes++,
    navigateTo: async (to: unknown) => void navigations.push(to),
  });
  return { calls, navigations, refreshes: () => refreshes };
}

const note = { id: 'n1', kind: 'note', title: 'Note' } as Item;
const folder = { id: 'f1', kind: 'folder', title: 'Folder' } as Item;

test('removing deletes each item at its route, then reads again once', async () => {
  const api = fake();
  await useItems().removeItems([note, folder]);
  assert.deepEqual(api.calls, [
    { url: '/api/notes/n1', method: 'DELETE', body: undefined },
    { url: '/api/folders/f1', method: 'DELETE', body: undefined },
  ]);
  assert.equal(api.refreshes(), 1);
});

test('a failed request reads nothing again', async () => {
  const api = fake({ fails: true });
  assert.equal(await useItems().removeItem(note), undefined);
  assert.equal(api.refreshes(), 0);
});

test('moving takes one id or several, the root without a parent', async () => {
  const api = fake();
  await useItems().moveItem('n1', 'f1');
  await useItems().moveItem(['n1', 'n2'], null);
  assert.deepEqual(
    api.calls.map((call) => call.body),
    [
      { ids: ['n1'], parent_id: 'f1' },
      { ids: ['n1', 'n2'], parent_id: undefined },
    ],
  );
});

test('the trash restores, purges one removal or empties', async () => {
  const api = fake();
  const items = useItems();
  await items.restoreItem('n1');
  await items.purgeItem('r1');
  await items.emptyTrash();
  assert.deepEqual(api.calls, [
    { url: '/api/restore', method: 'POST', body: { ids: ['n1'] } },
    { url: '/api/trash/r1', method: 'DELETE', body: undefined },
    { url: '/api/trash', method: 'DELETE', body: undefined },
  ]);
  assert.equal(api.refreshes(), 3);
});

test('renaming a tag on its page goes to the new one', async () => {
  const api = fake({ path: '/tags/old%20tag', answer: 'new tag' });
  await useItems().renameTag('old tag');
  assert.deepEqual(api.calls, [{ url: '/api/tags/old%20tag', method: 'PATCH', body: { value: 'new tag' } }]);
  assert.deepEqual(api.navigations, ['/tags/new%20tag']);
  assert.equal(api.refreshes(), 1);
});

test('renaming a tag elsewhere stays there, and an unchanged name sends nothing', async () => {
  const api = fake({ answer: 'new' });
  await useItems().renameTag('old');
  assert.deepEqual(api.navigations, []);
  await useItems().renameTag('new');
  assert.equal(api.calls.length, 1);
});

test('a tag is deleted only once confirmed, its page left for the home', async () => {
  const refused = fake({ answer: false });
  await useItems().removeTag('a');
  assert.deepEqual(refused.calls, []);

  const api = fake({ path: '/tags/a', answer: true });
  await useItems().removeTag('a');
  assert.deepEqual(api.calls, [{ url: '/api/tags/a', method: 'DELETE', body: undefined }]);
  assert.deepEqual(api.navigations, ['/']);
});

test('a failed tag change stays on the page and reads nothing again', async () => {
  const api = fake({ path: '/tags/a', answer: 'b', fails: true });
  await useItems().renameTag('a');
  assert.deepEqual(api.navigations, []);
  assert.equal(api.refreshes(), 0);
});

const labels = (menu: { label?: string }[][]) => menu.map((group) => group.map((entry) => entry.label));

test('a note offers to pin or unpin it, then to delete it', () => {
  fake();
  assert.deepEqual(labels(useItems().itemMenu(note)), [['items.pin'], ['common.delete']]);
  fake({ pinned: ['n1'] });
  assert.deepEqual(labels(useItems().itemMenu(note)), [['items.unpin'], ['common.delete']]);
});

test('a folder offers what goes in it, favorite, rename, then delete', () => {
  fake({ pinned: ['f1'] });
  assert.deepEqual(labels(useItems().itemMenu(folder)), [
    ['common.newNote', 'common.newFolder'],
    ['items.removeFavorite', 'common.rename'],
    ['common.delete'],
  ]);
});

test('pinning from the menu patches the item at its route', async () => {
  const api = fake();
  const [[pin]] = useItems().itemMenu(folder).slice(1);
  await pin!.onSelect!(new Event('select'));
  assert.deepEqual(api.calls, [{ url: '/api/folders/f1', method: 'PATCH', body: { pinned: true } }]);
});
