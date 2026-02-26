export const useDataStore = defineStore('data', () => {
  /* Define – only serializable state (no Map, no class instances) */

  const _raw = ref<DataSchema[]>([]);
  const hasLoaded = ref<boolean>(false);

  /* Single computed → one buildFromRaw per _raw change, no DataCloneError */
  const _derived = computed(() => handleStoreUpdate(_raw.value));
  const map = computed(() => _derived.value.map);
  const tree = computed(() => _derived.value.tree);
  const data = computed(() => _derived.value.data);
  const tags = computed(() => _derived.value.tags);

  const trash = computed(() => _derived.value.trash);
  const trashAll = computed(() => _derived.value.trashAll);
  const trashCount = computed(() => _derived.value.trash.length);

  const popularTags = computed(() =>
    Array.from(tags.value.values())
      .sort((a, b) => b.count - a.count)
      .slice(0, 5)
  );

  function handleRead(data: DataSchema[]) {
    if (!hasLoaded.value) hasLoaded.value = true;
    _raw.value = data;
    return _raw.value;
  }

  function handleReadId(data: DataSchema) {
    _raw.value = [..._raw.value, data];
    return _raw.value;
  }

  function handleUpdate(data: DataSchema | DataSchema[]) {
    const d = Array.isArray(data) ? data : [data];
    _raw.value = [..._raw.value, ...d];
    return _raw.value;
  }

  function handleUpdateId(id: string, data: Partial<DataSchema>) {
    const index = _raw.value.findIndex((n) => n.id === id);
    if (index === -1) return;

    const current = _raw.value[index];
    const unchecked = defuDedupArrays(data, current);

    const updated = dataParams.safeParse(unchecked);
    if (!updated.success) return;

    _raw.value = _raw.value.map((item, i) =>
      i === index ? updated.data : item
    );
    return _raw.value;
  }

  function handleDelete() {
    _raw.value = [];
    return _raw.value;
  }

  function handleDeleteId(id: string | string[]) {
    const ids = Array.isArray(id) ? id : [id];

    const filtered = _raw.value.filter((n) => !ids.includes(n.id));
    _raw.value = [...filtered];

    return _raw.value;
  }

  function handleSoftDelete(ids: string[]) {
    const now = new Date();
    _raw.value = _raw.value.map((item) =>
      ids.includes(item.id) && !item.deletedAt
        ? { ...item, deletedAt: now }
        : item
    );
    return _raw.value;
  }

  function handleRestore(ids: string[]) {
    _raw.value = _raw.value.map((item) =>
      ids.includes(item.id) ? { ...item, deletedAt: null } : item
    );
    return _raw.value;
  }

  function handlePurge(ids: string[]) {
    _raw.value = _raw.value.filter((item) => !ids.includes(item.id));
    return _raw.value;
  }

  function handleStoreUpdate(items: DataSchema[]) {
    // Separate active from soft-deleted items
    const activeItems = items.filter((item) => !item.deletedAt);
    const deletedItems = items.filter((item) => !!item.deletedAt);

    const newMap: Lookup = new Map();
    const newTags: MappedTags = new Map();
    const rootFolder = new Folder({
      id: 'root',
      createdAt: new Date(),
      updatedAt: new Date(),
      deletedAt: null,
      path: [],
      type: ItemType.FOLDER,
      title: 'Root',
      childrenIds: [],
    });

    newMap.set(rootFolder.id, rootFolder);
    const newData: StoredData = {
      notes: [],
      folders: [rootFolder],
    };

    for (const item of activeItems) {
      const handlers = {
        [ItemType.FOLDER]: () => {
          const props = folderParams.parse(item);
          const folder = new Folder(props);
          newData.folders.push(folder);
          return folder;
        },
        [ItemType.NOTE]: () => {
          const props = noteParams.parse(item);
          const note = new Note(props);
          newData.notes.push(note);

          if (note.ancestor === null) {
            note.path = ['root'];
          }

          extractTags({ note, tags: newTags });
          return note;
        },
      };

      const handler = handlers[item.type];
      if (!handler) continue;

      const instance = handler();

      if (instance.ancestor === 'root') {
        rootFolder.childrenIds.push(instance.id);
      }

      newMap.set(instance.id, instance);
    }

    const newTree = rebuildTree(newData.folders);

    // Build trash: show root-level deleted items
    // An item is a trash root if its parent is NOT deleted,
    // or if it was deleted BEFORE its parent (independent deletion)
    const deletedMap = new Map(deletedItems.map((d) => [d.id, d]));
    const trashRoots = deletedItems.filter((item) => {
      const parentId = item.path.at(-1);
      if (!parentId || !deletedMap.has(parentId)) return true;
      const parent = deletedMap.get(parentId)!;
      return (
        new Date(item.deletedAt!).getTime() <
        new Date(parent.deletedAt!).getTime()
      );
    });

    return {
      map: newMap,
      data: newData,
      tags: newTags,
      tree: newTree,
      trash: trashRoots,
      trashAll: deletedItems,
    };
  }

  function getSnapshot(): DataSchema[];
  function getSnapshot(id: string): DataSchema | undefined;
  function getSnapshot(id?: string): DataSchema | DataSchema[] | undefined {
    return id ? _raw.value.find((n) => n.id === id) : _raw.value;
  }

  return {
    tree,
    map,
    data,
    tags,
    popularTags,
    trash,
    trashAll,
    trashCount,
    hasLoaded,
    getSnapshot,
    loadData: handleRead,
    read: handleRead,
    readById: handleReadId,
    update: handleUpdate,
    updateById: handleUpdateId,
    delete: handleDelete,
    deleteById: handleDeleteId,
    softDelete: handleSoftDelete,
    restore: handleRestore,
    purge: handlePurge,
  };
});
