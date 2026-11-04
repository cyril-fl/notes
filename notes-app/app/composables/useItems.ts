import type { ContextMenuItem } from '@nuxt/ui';
import type { Item } from '~/types/items';

const CYCLE = 'Un dossier ne peut pas aller dans lui-même ou ses sous-dossiers.';
const PARENT_IN_TRASH = "Le dossier parent est dans la corbeille : restaure-le d'abord.";

/** Every change to the items: the views read again what they show afterwards. */
export function useItems() {
  const api = useApi();
  const { prompt } = useDialogs();

  // On failure the toast already told the user.
  async function act<T>(request: () => Promise<T>): Promise<T | undefined> {
    try {
      const result = await request();
      await refreshNuxtData();
      return result;
    } catch {
      return undefined;
    }
  }

  async function createFolder(parentId?: string) {
    const title = await prompt({ title: 'Nouveau dossier', placeholder: 'Nom du dossier', action: 'Créer' });
    if (!title) return;
    const body = { title, parent_id: parentId };
    return act(() => api<Item>('/api/folders', { method: 'POST', body }));
  }

  async function renameFolder(folder: Item) {
    const title = await prompt({ title: 'Renommer le dossier', value: folder.title, action: 'Renommer' });
    if (!title || title === folder.title) return;
    return act(() => api<Item>(`/api/folders/${folder.id}`, { method: 'PATCH', body: { title } }));
  }

  function removeItem(item: Pick<Item, 'id' | 'kind'>) {
    const route = item.kind === 'folder' ? 'folders' : 'notes';
    return act(() => api(`/api/${route}/${item.id}`, { method: 'DELETE' }));
  }

  /** A null parent is the root. */
  function moveItem(id: string, parentId: string | null) {
    const body = { ids: [id], parent_id: parentId ?? undefined };
    return act(() => api<Item[]>('/api/move', { method: 'POST', body, conflict: CYCLE }));
  }

  function restoreItem(id: string) {
    return act(() => api<Item[]>('/api/restore', { method: 'POST', body: { ids: [id] }, conflict: PARENT_IN_TRASH }));
  }

  function emptyTrash() {
    return act(() => api('/api/trash', { method: 'DELETE' }));
  }

  /** What right-clicking an item offers. */
  function itemMenu(item: Item): ContextMenuItem[][] {
    const remove = { label: 'Supprimer', icon: 'i-lucide-trash-2', color: 'error' as const, onSelect: () => removeItem(item) };
    if (item.kind === 'note') return [[remove]];

    return [
      [
        { label: 'Nouvelle note', icon: 'i-lucide-square-pen', onSelect: () => navigateTo(newNotePath(item.id)) },
        { label: 'Nouveau dossier', icon: 'i-lucide-folder-plus', onSelect: () => createFolder(item.id) },
      ],
      [{ label: 'Renommer', icon: 'i-lucide-pencil', onSelect: () => renameFolder(item) }],
      [remove],
    ];
  }

  return { createFolder, renameFolder, removeItem, moveItem, restoreItem, emptyTrash, itemMenu };
}

export function newNotePath(parentId?: string) {
  return parentId ? `/notes/new?parent=${parentId}` : '/notes/new';
}

/** Items a folder holds, the root without an id. */
export function useChildren(parentId?: string) {
  const api = useApi();
  const query = parentId ? { parent: parentId } : {};
  return useLazyAsyncData(`items:${parentId ?? 'root'}`, () => api<Item[]>('/api/items', { query }), {
    default: () => [] as Item[],
  });
}
