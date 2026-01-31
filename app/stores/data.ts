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

  const popularTags = computed(() =>
    Array.from(tags.value.values())
      .sort((a, b) => b.count - a.count)
      .slice(0, 5)
  );

  /* Methods */
  const handleRead = (data: DataSchema[]) => {
    if (!hasLoaded.value) hasLoaded.value = true;
    _raw.value = data;
    return _raw.value;
  };

  const handleReadId = (data: DataSchema) => {
    _raw.value = [..._raw.value, data];
    return _raw.value;
  };

  const handleUpdate = (data: DataSchema | DataSchema[]) => {
    const d = Array.isArray(data) ? data : [data];
    _raw.value = [..._raw.value, ...d];
    return _raw.value;
  };

  const handleUpdateId = (id: string, data: Partial<DataSchema>) => {
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
  };

  const handleDelete = () => {
    _raw.value = [];
    return _raw.value;
  };

  const handleDeleteId = (id: string | string[]) => {
    const ids = Array.isArray(id) ? id : [id];

    const filtered = _raw.value.filter((n) => !ids.includes(n.id));
    _raw.value = [...filtered];

    return _raw.value;
  };

  function handleStoreUpdate(items: DataSchema[]) {
    const newMap: Lookup = new Map();
    const newTags: MappedTags = new Map();
    const newData: StoredData = {
      notes: [],
      folders: [],
    };

    for (const item of items) {
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
          extractTags({ note, tags: newTags });
          return note;
        },
      };

      const handler = handlers[item.type];
      if (!handler) continue;

      const instance = handler();
      newMap.set(instance.id, instance);
    }

    const newTree = rebuildTree(newData.folders);

    return { map: newMap, data: newData, tags: newTags, tree: newTree };
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
    hasLoaded,
    onRead: handleRead,
    onReadId: handleReadId,
    onUpdate: handleUpdate,
    onUpdateId: handleUpdateId,
    onDelete: handleDelete,
    onDeleteId: handleDeleteId,
    getSnapshot,
  };
});
