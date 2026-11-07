import { execFileSync } from 'node:child_process';
import { request, type APIRequestContext } from '@playwright/test';
import { API, APP, SESSION } from '../../playwright.config';

/** A folder of notes made through the API for one test, deleted for good afterwards: titles are unique per run. */
export async function sandbox() {
  const api: APIRequestContext = await request.newContext({ baseURL: API, extraHTTPHeaders: { Origin: APP }, storageState: SESSION });
  const prefix = `e2e-${Date.now()}`;
  const folder = await (await api.post('/api/folders', { data: { title: prefix } })).json();

  return {
    api,
    prefix,
    folder,
    async note(content: string) {
      const response = await api.post('/api/notes', { data: { content, parent_id: folder.id } });
      return (await response.json()) as { id: string; title: string };
    },
    async content(id: string) {
      return ((await (await api.get(`/api/notes/${id}`)).json()) as { content: string }).content;
    },
    /** Leaves nothing behind, not even in the trash. */
    async dispose() {
      await api.delete(`/api/folders/${folder.id}`);
      await api.delete(`/api/trash/${folder.id}`);
      await api.dispose();
    },
  };
}

/**
 * An account of its own, for what would break the shared e2e account: a real purge, a password change, another
 * account's data. Made by the CLI in the API's container, deleted for good with its data afterwards.
 * shortcut: needs the containers of `make up` (note-api, note-postgres), give the API an admin route if E2E_API ever points elsewhere.
 */
export async function throwaway() {
  const name = `e2e-${Date.now()}`;
  const password = 'throwaway-password';
  execFileSync('docker', ['exec', '-i', 'note-api', '/notes', 'user', 'add', name, '--password-stdin'], { input: `${password}\n` });

  const api = await request.newContext({ baseURL: API, extraHTTPHeaders: { Origin: APP } });
  const login = await api.post('/api/login', { data: { name, password } });
  if (!login.ok()) throw new Error(`throwaway login: ${login.status()} ${await login.text()}`);

  return {
    name,
    password,
    api,
    storageState: await api.storageState(),
    async dispose() {
      await api.dispose();
      // The users FKs cascade to the items, tags, files and sessions of the account.
      execFileSync('docker', ['exec', '-i', 'note-postgres', 'sh', '-c', `psql -q -v ON_ERROR_STOP=1 -v name=${name} -U "$POSTGRES_USER" "$POSTGRES_DB"`], {
        input: "DELETE FROM users WHERE name = :'name'",
      });
    },
  };
}
