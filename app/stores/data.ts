import { defu } from 'defu';

export const useDataStore = defineStore('data', () => {
  /* Define */
  const _raw = ref<DataSchema[]>([]);

  const tree = ref<Tree>({});
  const map = shallowRef<Lookup>(new Map());
  const data = shallowRef<Data>({
    notes: [],
    folders: [],
  });
  const tags = ref<MappedTags>(new Map());

  const hasLoaded = ref<boolean>(false);

  /* Data */
  const popularTags = computed(() =>
    Array.from(tags.value.values())
      .sort((a, b) => b.count - a.count)
      .slice(0, 5)
  );

  /* Methods */
  const handleRead = (data: DataSchema[]) => {
    if (!hasLoaded.value) {
      hasLoaded.value = true;
    }

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
    const unchecked = defu(data, current);

    const updated = dataParams.safeParse(unchecked);

    if (!updated.success) return;

    _raw.value[index] = updated.data;

    return _raw.value;
  };

  const handleDelete = () => {
    _raw.value = [];

    console.log('After delete, raw data:', _raw.value);
    return _raw.value;
  };

  const handleDeleteId = (id: string) => {
    _raw.value = _raw.value.filter((n) => n.id !== id);

    return _raw.value;
  };

  const handleStoreUpdate = (items: DataSchema[]) => {
    const newMap: Lookup = new Map();
    const newTags: MappedTags = new Map();
    const newData: Data = {
      notes: [],
      folders: [],
    };

    for (const item of items) {
      let instance: Note | Folder | null = null;

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

      instance = handler();

      newMap.set(instance.id, instance);
    }

    map.value = newMap;
    tags.value = newTags;
    data.value = newData;
  };

  function getSnapshot(): DataSchema[];
  function getSnapshot(id: string): DataSchema | undefined;
  function getSnapshot(id?: string): DataSchema | DataSchema[] | undefined {
    return id
      ? structuredClone(_raw.value.find((n) => n.id === id))
      : structuredClone(_raw.value);
  }

  /* Lifecycle */
  watch(
    _raw,
    (newValue) => {
      console.log('Raw data changed, updating store...', newValue);
      handleStoreUpdate(newValue);
    },
    { immediate: true, deep: true }
  );

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
