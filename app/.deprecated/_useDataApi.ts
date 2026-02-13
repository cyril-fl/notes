export function useDataApi() {
  /* Define */
  const store = useDataStore();
  const { getById, getRelatedIds, checkPathValidity } = useDataUtils();

  const error = useState<Error | null>('error', () => null);
  const isLoading = useState('isLoading', () => false);

  const _fetch = useFetchApi({ isLoading, error });

  /* Data */

  /* Methods */
  // CRUD Basic Operations
  async function handleCreate(
    data: DraftData
  ): Promise<DataSchema | undefined> {
    const mock: DataSchema = {
      ...data,
      id: `mock-${generateId(data.type)}`,
      createdAt: new Date(),
      updatedAt: new Date(),
    };

    store.onUpdate(mock);

    const result = await _fetch<DataSchema | undefined, DraftData>(
      {
        url: 'data/create',
        method: HTTPMethod.POST,
        body: data,
      },
      'CREATE'
    )
      .onSuccess((res) => {
        if (!res.data) throw new Error('No data returned from create API');
        store.onUpdateId(mock.id, res.data);
      })
      .onError((_err) => {
        store.onDeleteId(mock.id);
      });

    if (!result.ok || !result.response.data) return;

    return result.response.data;
  }

  async function handleRead(): Promise<DataSchema[] | undefined> {
    const result = await _fetch<DataSchema[]>(
      {
        url: 'data/read',
        method: HTTPMethod.GET,
      },
      'READ'
    );

    if (!result.ok) return;

    store.onRead(result.response.data);
    await nextTick();

    return result.response.data;
  }

  async function handleUpdate(
    id: string,
    data: Partial<DraftData>
  ): Promise<DataSchema | undefined> {
    const snapshot = store.getSnapshot(id);
    const current = getById(id);

    if (!current || !snapshot) {
      console.warn(`[DATA API] - UPDATE  No item found with id ${id}`);
      return;
    }

    const optimisticUpdate: Partial<DataSchema> = {
      ...data,
      updatedAt: new Date(),
    };

    store.onUpdateId(id, optimisticUpdate);

    const result = await _fetch<DataSchema, Partial<DraftData>>(
      {
        url: `data/update/${id}`,
        method: HTTPMethod.PATCH,
        body: data,
      },
      'UPDATE'
    )
      .onSuccess((res) => {
        if (!res.data) throw new Error('No data returned from update API');
        store.onUpdateId(id, res.data);
      })
      .onError(() => {
        store.onUpdateId(id, snapshot);
      });

    if (!result.ok) return;

    await nextTick();

    return result.response.data;
  }

  async function handleDelete(): Promise<boolean> {
    const snapshot = store.getSnapshot();

    store.onDelete();

    const result = await _fetch<boolean>(
      {
        url: `data/delete`,
        method: HTTPMethod.DELETE,
      },
      'DELETE'
    ).onError((_err) => {
      console.warn(
        '[DATA API] - DELETE Error occurred, restoring previous state'
      );
      store.onRead(snapshot);
    });

    if (!result.ok) return false;

    await nextTick();

    return result.response.data;
  }

  async function handleDeleteId(id: string): Promise<boolean> {
    const snapshot = store.getSnapshot(id);

    if (!snapshot) {
      console.warn(`[DATA API] - DELETE No item found with id ${id}`);
      return false;
    }

    const relatedIds = getRelatedIds(id, { includeSelf: true });

    store.onDeleteId(relatedIds);

    const result = await _fetch<boolean>(
      {
        url: `data/delete/${relatedIds.join(',')}`,
        method: HTTPMethod.DELETE,
      },
      'DELETE'
    ).onError((_err) => {
      store.onReadId(snapshot);
    });

    if (!result.ok) return false;

    await nextTick();

    return result.response.data;
  }

  // CRUD Advanced Operations
  async function handleCreateFolderInFolder({
    folder,
    title,
  }: CreateFolderInFolderParameters) {
    const result = await handleCreate({
      path: folder.path,
      type: ItemType.FOLDER,
      title,
      childrenIds: [],
    });
    if (!result) return;

    const children = new Set([...folder.childrenIds, result.id]);

    await handleUpdate(folder.id, {
      childrenIds: Array.from(children),
    });
  }

  async function handleCreateNoteInFolder({
    folder,
    content,
  }: CreateNoteInFolderParameters) {
    const { id, path, childrenIds } = folder;

    // Todo utiliser un genre de useAlert
    if (!checkPathValidity(path, { throwError: true })) return;

    const result = await handleCreate({
      type: ItemType.NOTE,
      content,
      path,
    });

    if (!result) return;

    const children = new Set([...childrenIds, result.id]);

    await handleUpdate(id, {
      childrenIds: Array.from(children),
    });
  }

  return {
    // State
    error,
    isLoading,

    // CRUD
    handleCreate,
    handleRead,
    handleUpdate,
    handleDelete,
    handleDeleteId,
    // CRUD Advanced
    handleCreateFolderInFolder,
    handleCreateNoteInFolder,
  };
}
