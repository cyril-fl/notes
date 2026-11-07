import type { ContextMenuItem } from '@nuxt/ui';
import type { FolderPath, Item, Pinned } from '~/types/items';

// The part of the API's message saying a folder name is taken (domain.ErrFolderExists).
const TAKEN = 'already exists here';

/** Every change to the items: the views read again what they show afterwards. */
export function useItems() {
  const api = useApi();
  const route = useRoute();
  const { prompt, confirm } = useDialogs();
  const { icons } = useIcons();
  const { t } = useI18n();

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
    const title = await prompt({
      title: t('common.newFolder'),
      placeholder: t('items.folderName'),
      action: t('common.create'),
    });
    if (!title) return;
    const body = { title, parent_id: parentId };
    return act(() => api<Item>('/api/folders', { method: 'POST', body, conflict: t('items.nameTaken') }));
  }

  async function renameFolder(folder: Item) {
    const title = await prompt({ title: t('items.renameFolder'), value: folder.title, action: t('common.rename') });
    if (!title || title === folder.title) return;
    return act(() =>
      api<Item>(`/api/folders/${folder.id}`, { method: 'PATCH', body: { title }, conflict: t('items.nameTaken') }),
    );
  }

  function removeItem(item: Pick<Item, 'id' | 'kind'>) {
    return removeItems([item]);
  }

  /** The API removes one item per call: they go together, read again once. */
  function removeItems(items: Pick<Item, 'id' | 'kind'>[]) {
    return act(() =>
      Promise.all(
        items.map((item) =>
          api(`/api/${item.kind === 'folder' ? 'folders' : 'notes'}/${item.id}`, { method: 'DELETE' }),
        ),
      ),
    );
  }

  /** One id or several; a null parent is the root. */
  function moveItem(ids: string | string[], parentId: string | null) {
    const body = { ids: [ids].flat(), parent_id: parentId ?? undefined };
    return act(() =>
      api<Item[]>('/api/move', {
        method: 'POST',
        body,
        conflict: { [TAKEN]: t('items.nameTaken'), 'into itself': t('items.cycle') },
      }),
    );
  }

  function restoreItem(id: string) {
    return act(() =>
      api<Item[]>('/api/restore', {
        method: 'POST',
        body: { ids: [id] },
        conflict: { [TAKEN]: t('items.nameTaken'), 'in the trash': t('items.parentInTrash') },
      }),
    );
  }

  /** Pins a note, or makes a folder a favorite. */
  function pinItem(item: Pick<Item, 'id' | 'kind'>, pinned: boolean) {
    const resource = item.kind === 'folder' ? 'folders' : 'notes';
    return act(() => api(`/api/${resource}/${item.id}`, { method: 'PATCH', body: { pinned } }));
  }

  async function renameTag(value: string) {
    const next = await prompt({ title: t('items.renameTag'), value, action: t('common.rename') });
    if (!next || next === value) return;
    await changeTag(value, { method: 'PATCH', body: { value: next }, conflict: t('items.tagTaken') }, tagPath(next));
  }

  /** Asks first: the tag is also erased from the text of its notes, and no trash keeps it. */
  async function removeTag(value: string) {
    const confirmed = await confirm({
      title: t('items.deleteTag', { value }),
      description: t('items.deleteTagDescription'),
      action: t('common.delete'),
    });
    if (confirmed) await changeTag(value, { method: 'DELETE' }, '/');
  }

  /** On the tag's page, leaves it for next before reading again: the old page would only find the tag gone. */
  async function changeTag(value: string, options: Parameters<typeof api>[1], next: string) {
    try {
      await api(`/api${tagPath(value)}`, options);
    } catch {
      return; // The toast already told the user.
    }
    if (route.path === tagPath(value)) await navigateTo(next, { replace: true });
    await refreshNuxtData();
  }

  function tagMenu(value: string): ContextMenuItem[][] {
    return [
      [{ label: t('common.rename'), icon: icons.rename, onSelect: () => renameTag(value) }],
      [{ label: t('common.delete'), icon: icons.delete, color: 'error' as const, onSelect: () => removeTag(value) }],
    ];
  }

  function emptyTrash() {
    return act(() => api('/api/trash', { method: 'DELETE' }));
  }

  /** Deletes one removal for good, with what went with it. */
  function purgeItem(id: string) {
    return act(() => api(`/api/trash/${id}`, { method: 'DELETE' }));
  }

  /** What right-clicking an item offers. */
  function itemMenu(item: Item): ContextMenuItem[][] {
    const remove = {
      label: t('common.delete'),
      icon: icons.delete,
      color: 'error' as const,
      onSelect: () => removeItem(item),
    };
    if (item.kind === 'note') {
      const pinned = isPinned(item.id);
      const pin = {
        label: pinned ? t('items.unpin') : t('items.pin'),
        icon: pinned ? icons.unpin : icons.pin,
        onSelect: () => pinItem(item, !pinned),
      };
      return [[pin], [remove]];
    }

    const favorite = isPinned(item.id);

    return [
      [
        { label: t('common.newNote'), icon: icons.noteAdd, onSelect: () => navigateTo(newNotePath(item.id)) },
        { label: t('common.newFolder'), icon: icons.folderAdd, onSelect: () => createFolder(item.id) },
      ],
      [
        {
          label: favorite ? t('items.removeFavorite') : t('items.addFavorite'),
          icon: favorite ? icons.unfavorite : icons.favorite,
          onSelect: () => pinItem(item, !favorite),
        },
        { label: t('common.rename'), icon: icons.rename, onSelect: () => renameFolder(item) },
      ],
      [remove],
    ];
  }

  return {
    createFolder,
    renameFolder,
    removeItem,
    removeItems,
    moveItem,
    restoreItem,
    pinItem,
    emptyTrash,
    purgeItem,
    itemMenu,
    renameTag,
    removeTag,
    tagMenu,
  };
}

/** The pinned notes and favorite folders, read once for the sidebar, the menus and the editor. */
export function usePinned() {
  const api = useApi();
  return useLazyAsyncData('pinned', () => api<Pinned[]>('/api/pinned'), { default: () => [] as Pinned[] });
}

function isPinned(id: string) {
  return useNuxtData<Pinned[]>('pinned').data.value?.some((note) => note.id === id) ?? false;
}

/** Stands for the root in a select, which refuses '' as a value; the API takes no parent for it. */
export const ROOT = 'root';

/** Where items can go, for a select: the root, then every folder by its path but the excluded ones. */
export function useDestinations(exclude: () => string[] = () => []) {
  const api = useApi();
  const { t } = useI18n();
  const { data: folders } = useLazyAsyncData('folder-paths', () => api<FolderPath[]>('/api/folders'), {
    default: () => [] as FolderPath[],
  });
  return computed(() => [
    { label: t('common.root'), value: ROOT },
    ...folders.value
      .filter((folder) => !exclude().includes(folder.id))
      .map((folder) => ({ label: folder.path, value: folder.id })),
  ]);
}

/** The folder the page shows, where a new note or folder goes; undefined at the root and elsewhere. */
export function currentFolder(route: { path: string; params: Record<string, unknown> }) {
  return route.path.startsWith('/folders/') ? (route.params.id as string) : undefined;
}

/** The page of a tag; the API's route is the same under /api. */
export function tagPath(value: string) {
  return `/tags/${encodeURIComponent(value)}`;
}

// template: the index of a template in the account's snippets.
export function newNotePath(parentId?: string, template?: number) {
  return { path: '/notes/new', query: { parent: parentId, template } };
}

/** Items a folder holds, the root without an id. */
export function useChildren(parentId?: string) {
  const api = useApi();
  const query = parentId ? { parent: parentId } : {};
  return useLazyAsyncData(`items:${parentId ?? 'root'}`, () => api<Item[]>('/api/items', { query }), {
    default: () => [] as Item[],
  });
}
