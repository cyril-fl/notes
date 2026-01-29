export function useDataApi() {
  /* Define */
  const store = useDataStore();
  const { getById } = useDataUtils();

  const error = useState<Error | null>('error', () => null);
  const isLoading = useState('isLoading', () => false);

  const _fetch = useFetchApi({ isLoading, error });

  /* Data */

  /* Methods */
  // CRUD Operations
  const handleCreate = async (
    data: DraftData
  ): Promise<DataSchema | undefined> => {
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
        method: 'POST',
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
  };

  const handleRead = async (): Promise<DataSchema[] | undefined> => {
    const result = await _fetch<DataSchema[]>(
      {
        url: 'data/read',
        method: 'GET',
      },
      'READ'
    );

    if (!result.ok) return;

    store.onRead(result.response.data);
    await nextTick();

    return result.response.data;
  };

  const handleUpdate = async (
    id: string,
    data: Partial<DraftData>
  ): Promise<DataSchema | undefined> => {
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
        method: 'POST',
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
  };

  const handleDelete = async (): Promise<boolean> => {
    const snapshot = store.getSnapshot();

    store.onDelete();

    const result = await _fetch<boolean>(
      {
        url: `data/delete`,
        method: 'POST',
      },
      'DELETE'
    ).onError((_err) => {
      store.onRead(snapshot);
    });

    if (!result.ok) return false;

    await nextTick();

    return result.response.data;
  };

  const handleDeleteId = async (id: string): Promise<boolean> => {
    const snapshot = store.getSnapshot(id);

    if (!snapshot) {
      console.warn(`[DATA API] - DELETE No item found with id ${id}`);
      return false;
    }

    store.onDeleteId(id);

    const result = await _fetch<boolean>(
      {
        url: `data/delete/${id}`,
        method: 'POST',
      },
      'DELETE'
    ).onError((_err) => {
      store.onReadId(snapshot);
    });

    if (!result.ok) return false;

    await nextTick();

    return result.response.data;
  };

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
  };
}
