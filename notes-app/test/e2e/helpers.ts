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
