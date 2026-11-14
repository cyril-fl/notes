import type { ContextMenuItem } from '@nuxt/ui';
import type { FolderPath, Item, ItemDetail, Pinned, SavedSearch, Tag, User } from '~/types/items';

// The part of the API's message saying a folder name is taken (domain.ErrFolderExists).
const TAKEN = 'already exists here';
// The part of the API's message saying the parent folder is gone or in the trash (ErrParentNotFound).
export const NO_PARENT = 'parent folder not found';

/** Every change to the items: the views read again what they show afterwards. */
export function useItems() {
  const api = useApi();
  const route = useRoute();
  const { prompt, confirm } = useDialogs();
  const { icons } = useIcons();
  const { t } = useI18n();
  const toast = useToast();
  const { has } = usePro();
  const { addSavedSearch } = useSavedSearches();

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
      action: t('common.save'),
      // A smart folder lives in the sidebar, not in a folder: offered at the root only.
      alternative:
        !parentId && has('savedSearches') ? { label: t('search.saved.add'), onSelect: addSavedSearch } : undefined,
    });
    if (!title) return;
    const body = { title, parent_id: parentId };
    return act(() =>
      api<Item>('/api/folders', {
        method: 'POST',
        body,
        conflict: t('items.nameTaken'),
        invalid: { [NO_PARENT]: t('items.parentInTrash') },
      }),
    );
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
        invalid: { [NO_PARENT]: t('items.parentInTrash') },
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

  /** Downloads the zip of these items with what they hold: the API sends it as an attachment, the page stays. */
  function exportItems(ids: string[]) {
    const query = new URLSearchParams(ids.map((id) => ['id', id]));
    window.location.assign(`${useRuntimeConfig().public.apiBase}/api/export?${query}`);
  }

  /** Hands the note's markdown to the system's share sheet, or to the clipboard where there is none. */
  async function shareItem(item: Item) {
    const note = await api<ItemDetail>(`/api/notes/${item.id}`).catch(() => undefined);
    if (!note) return; // The toast already told the user.
    const title = displayTitle(item.title, t('untitled'));
    const text = note.content ?? '';
    if (navigator.share) {
      // Cancelling the sheet rejects: nothing to say.
      await navigator.share({ title, text }).catch(() => undefined);
      return;
    }
    try {
      await navigator.clipboard.writeText(text);
      toast.add({ title: t('items.copied'), color: 'success' });
    } catch {
      // No clipboard either: the page is not a secure context (plain http).
      toast.add({ title: t('items.shareFailed'), color: 'error' });
    }
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
      return [
        [pin],
        [
          { label: t('items.share'), icon: icons.share, onSelect: () => shareItem(item) },
          { label: t('settings.export'), icon: icons.export, onSelect: () => exportItems([item.id]) },
        ],
        [remove],
      ];
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
        { label: t('settings.export'), icon: icons.export, onSelect: () => exportItems([item.id]) },
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
    exportItems,
    shareItem,
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
  // defer: the views that each ask for it at once share the pending request rather than cancelling and resending it.
  const { data: folders } = useLazyAsyncData('folder-paths', () => api<FolderPath[]>('/api/folders'), {
    default: () => [] as FolderPath[],
    dedupe: 'defer',
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

// snippet: the index of the account's snippet the note starts with.
export function newNotePath(parentId?: string, snippet?: number) {
  return { path: '/notes/new', query: { parent: parentId, snippet } };
}

/** Items a folder holds, the root without an id; only those of this kind when given. */
export function useChildren(parentId?: string, kind?: Item['kind']) {
  const api = useApi();
  const query = { parent: parentId, kind };
  const key = `items:${parentId ?? 'root'}${kind ? `:${kind}` : ''}`;
  return useLazyAsyncData(key, () => api<Item[]>('/api/items', { query }), {
    default: () => [] as Item[],
  });
}

// Children read at once: about 40 KB, where a folder of 48,000 notes was 9 MB in one answer.
const PAGE = 200;

/** Items a folder holds by pages, the root without an id: a hole stands for an item not read yet, load fills them. */
export function useFolderPages(parentId?: string) {
  const api = useApi();
  const loading = new Set<number>();

  async function read(page: number) {
    let total = 0;
    const items = await api<Item[]>('/api/items', {
      query: { parent: parentId, limit: PAGE, offset: page * PAGE },
      onResponse: ({ response }) => void (total = Number(response.headers.get('X-Total-Count'))),
    });
    return { page, items, total };
  }

  function place(all: (Item | undefined)[], { page, items, total }: Awaited<ReturnType<typeof read>>) {
    all.length = total;
    all.splice(page * PAGE, items.length, ...items);
  }

  // Read again after every change: the first page and those read since, the total as it is now.
  const { data, status } = useLazyAsyncData(
    `items:${parentId ?? 'root'}:pages`,
    async () => {
      const pages = [0];
      for (let page = 1; page * PAGE < data.value.length; page++) if (data.value[page * PAGE]) pages.push(page);
      const all: (Item | undefined)[] = [];
      for (const result of await Promise.all(pages.map(read))) place(all, result);
      return all;
    },
    { default: () => [] as (Item | undefined)[] },
  );

  /** Reads the pages holding the items from one index to another, both included, that are not read yet. */
  async function load(from: number, to: number) {
    for (let page = Math.floor(from / PAGE); page <= Math.floor(to / PAGE); page++) {
      if (data.value[page * PAGE] || loading.has(page)) continue;
      loading.add(page);
      try {
        const result = await read(page);
        const all = [...data.value];
        place(all, result);
        data.value = all;
      } catch {
        // The toast already told the user; scrolling there again tries again.
      } finally {
        loading.delete(page);
      }
    }
  }

  return { items: data, status, load };
}

/** Every tag of the account, with its id:a saved search's criteria match tags by id, never by their decrypted value. */
export function useTags() {
  const api = useApi();
  return useLazyAsyncData('tags-list', () => api<Tag[]>('/api/tags'), { default: () => [] as Tag[] });
}

/** Saved searches ("smart folders"): create/edit/delete/reorder all replace the whole list via PATCH /api/me,
 * exactly like the snippets of the settings page. */
export function useSavedSearches() {
  const api = useApi();
  const { t } = useI18n();
  const { icons } = useIcons();
  const { savedSearch, confirm } = useDialogs();

  async function save(searches: SavedSearch[]) {
    const me = await api<User>('/api/me', { method: 'PATCH', body: { saved_searches: searches } });
    useNuxtData<User>('me').data.value = me;
    return me.saved_searches;
  }

  function list() {
    return useNuxtData<User>('me').data.value?.saved_searches ?? [];
  }

  async function addSavedSearch() {
    const added = await savedSearch({ title: t('search.saved.add') });
    if (!added) return;
    await save([...list(), { ...added, id: crypto.randomUUID() }]);
  }

  async function editSavedSearch(search: SavedSearch) {
    const changed = await savedSearch({ title: t('search.saved.edit'), value: search });
    if (!changed) return;
    await save(list().map((s) => (s.id === search.id ? { ...changed, id: search.id } : s)));
  }

  async function removeSavedSearch(search: SavedSearch) {
    const confirmed = await confirm({
      title: t('search.saved.delete', { name: search.name }),
      action: t('common.delete'),
    });
    if (confirmed) await save(list().filter((s) => s.id !== search.id));
  }

  /** Drops what useDragItem carried onto this saved search, moving the dragged one right before it. */
  async function reorderSavedSearch(draggedId: string, targetId: string) {
    if (draggedId === targetId) return;
    const searches = [...list()];
    const from = searches.findIndex((s) => s.id === draggedId);
    const to = searches.findIndex((s) => s.id === targetId);
    if (from < 0 || to < 0) return;
    const [moved] = searches.splice(from, 1);
    searches.splice(to, 0, moved!);
    await save(searches);
  }

  function savedSearchMenu(search: SavedSearch): ContextMenuItem[][] {
    return [
      [{ label: t('common.edit'), icon: icons.rename, onSelect: () => editSavedSearch(search) }],
      [
        {
          label: t('common.delete'),
          icon: icons.delete,
          color: 'error' as const,
          onSelect: () => removeSavedSearch(search),
        },
      ],
    ];
  }

  return { addSavedSearch, editSavedSearch, removeSavedSearch, reorderSavedSearch, savedSearchMenu };
}
