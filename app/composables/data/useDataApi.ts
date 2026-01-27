type BodyRequest = Record<string, unknown> | undefined;
interface FetchOptions<T extends BodyRequest = undefined> {
  url: string;
  method: 'GET' | 'POST' | 'PUT' | 'DELETE';
  body?: T;
}

type FetchResult<R> =
  | { ok: true; response: ServerResponse<R> }
  | { ok: false; error: Error };

export function useDataApi() {
  /* Define */
  const store = useDataStore();
  const { getById } = useDataUtils();

  const error = useState<Error | null>('error', () => null);
  const isLoading = useState('isLoading', () => false);

  /* TODO

  - transforme fetch en builder avec callback on success et on error dazns un utils externe
  - rendre le code optimisete
  - gerer les todo depuis le server
  */

  /* Data */

  /* Methods */
  const _fetch = async <R, T extends BodyRequest = undefined>(
    options: FetchOptions<T>,
    ctx?: string
  ): Promise<FetchResult<R>> => {
    const prefix = ctx ? `[ELEMENT API] - ${ctx}` : `[ELEMENT API]`;

    try {
      if (isLoading.value) throw new Error('Already loading...');

      isLoading.value = true;
      error.value = null;

      const response = await $fetch<ServerResponse<R>>(`/api/${options.url}`, {
        method: options.method,
        body: options.body,
        headers: {},
      });

      console.info(prefix + ' response', response);

      return { ok: true, response };
    } catch (e) {
      const err = e instanceof Error ? e : new Error(String(e));
      error.value = err;

      console.error(prefix + ' error', err);

      return { ok: false, error: err };
    } finally {
      isLoading.value = false;
    }
  };

  // CRUD Operations
  const handleCreate = async (
    data: DraftData
  ): Promise<DataPublic | undefined> => {
    const result = await _fetch<DataSchema | undefined, DraftData>(
      {
        url: 'data/create',
        method: 'POST',
        body: data,
      },
      'CREATE_ELEMENT'
    );

    if (!result.ok) return;

    const response = result.response;

    if (!response.data) return;

    store.onUpdate(response.data);
    await nextTick();

    const get = getById(response.data.id);

    return get;
  };

  const handleRead = async (): Promise<DataSchema[] | undefined> => {
    const result = await _fetch<DataSchema[]>(
      {
        url: 'data/read',
        method: 'GET',
      },
      'READ_ELEMENTS'
    );

    if (!result.ok) return;

    store.onRead(result.response.data);
    await nextTick();

    return result.response.data;
  };

  const handleUpdate = async (
    id: string,
    data: Partial<DataPublic>
  ): Promise<DataSchema | undefined> => {
    const result = await _fetch<DataSchema, Partial<DraftData>>(
      {
        url: `data/update/${id}`,
        method: 'POST',
        body: data,
      },
      'UPDATE_ELEMENT'
    );

    if (!result.ok) return;

    console.log('[DATA API - UPDATE] result', result);
    store.onUpdateId(id, result.response.data);
    await nextTick();

    return result.response.data;
  };

  const handleDelete = async (): Promise<boolean> => {
    const result = await _fetch<boolean>(
      {
        url: `data/delete`,
        method: 'POST',
      },
      'DELETE_ELEMENTS'
    );

    console.log('[DATA API - DELETE] result', result);

    if (!result.ok) return false;

    store.onDelete();
    await nextTick();

    return result.response.data;
  };

  const handleDeleteId = async (id: string): Promise<boolean> => {
    const result = await _fetch<boolean>(
      {
        url: `data/delete/${id}`,
        method: 'POST',
      },
      'DELETE_ELEMENT'
    );

    console.log('[DATA API - DELETE ID] result', result);

    if (!result.ok) return false;

    store.onDeleteId(id);
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
