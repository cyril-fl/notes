import { request, type APIRequestContext } from '@playwright/test';

const API = 'http://localhost:8080';

/** A folder of notes made through the API for one test, trashed afterwards: titles are unique per run. */
export async function sandbox() {
  const api: APIRequestContext = await request.newContext({ baseURL: API, extraHTTPHeaders: { Origin: 'http://localhost:3000' } });
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
    async dispose() {
      await api.delete(`/api/folders/${folder.id}`);
      await api.dispose();
    },
  };
}
