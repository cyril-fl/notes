import type { CreateFolder, CreateNote } from '~/types/data';

interface GetRelatedOptions {
  includeSelf: boolean;
}

interface CheckPathValidityOptions {
  throwError: boolean;
}

interface TypeOptions<T extends ItemType> {
  types: T;
}

enum CRUD {
  CREATE = 'CREATE',
  READ = 'READ',
  UPDATE = 'UPDATE',
  DELETE = 'DELETE',
}

export function useDataApi() {
  const error = useState<Error | null>('error', () => null);
  const isLoading = useState('isLoading', () => false);

  const fetch = useFetchApi({ isLoading, error });

  return {
    error,
    isLoading,
    fetch,
  };
}

export function useDataUtils() {
  const store = useDataStore();
  const { map, tree } = storeToRefs(store);

  function getById(id: string): Data | undefined;
  function getById<T extends ItemType>(
    id: string,
    options: TypeOptions<T>
  ): ItemByType<T> | undefined;
  function getById(
    id: string,
    options?: { types?: ItemType }
  ): Data | undefined {
    const result = map.value.get(id);

    if (!result) return;
    const desiredTypes = options?.types;

    const isTypeAsked = !!desiredTypes;
    const isTypeNotMatching = result.type !== desiredTypes;

    return !(isTypeAsked && isTypeNotMatching) ? result : undefined;
  }

  function getRelatedIds(
    id: string,
    options?: Partial<GetRelatedOptions>
  ): string[] {
    const includeSelf = !!options?.includeSelf;

    const crawlChildren = (id: string) => {
      const relatedIds = [id];
      const result = map.value.get(id);

      if (result && result.type === ItemType.FOLDER) {
        for (const child of result.childrenIds) {
          relatedIds.push(...crawlChildren(child));
        }
      }
      return relatedIds;
    };

    const selfAndRelatedIds = crawlChildren(id);

    return includeSelf
      ? selfAndRelatedIds
      : selfAndRelatedIds.filter((child) => child !== id);
  }

  function getChildrenCountByType<T extends ItemType>(
    id: string,
    options: TypeOptions<T>
  ): number | undefined {
    const result = map.value.get(id);
    if (!result || result.type !== ItemType.FOLDER) return undefined;

    const childrenIds = [];

    for (const childId of result.childrenIds) {
      const child = map.value.get(childId);
      if (!child || child.type !== options.types) continue;

      childrenIds.push(childId);
    }

    return childrenIds.length;
  }

  function checkPathValidity(
    path: string[],
    options?: Partial<CheckPathValidityOptions>
  ): boolean {
    const throwError = !!options?.throwError;

    // Map
    const isEverySegmentValid = path.every((segment) => {
      const result = map.value.get(segment);
      if (!result) return false;
      if (result.type !== ItemType.FOLDER) return false;
      return true;
    });

    if (throwError && !isEverySegmentValid) {
      throw createError({
        statusCode: 400,
        statusMessage: 'Invalid path',
      });
    }

    const isPathInTree = exploreTree({
      tree: tree.value,
      path,
      options: {
        asBoolean: true,
      },
    });

    if (throwError && !isPathInTree) {
      throw createError({
        statusCode: 400,
        statusMessage: 'Path not found',
      });
    }

    return isEverySegmentValid && isPathInTree;
  }

  return {
    getById,
    getRelatedIds,
    getChildrenCountByType,
    checkPathValidity,
  };
}

export function useDataActions() {
  const store = useDataStore();
  const { fetch: _fetch } = useDataApi();
  const { getById, getRelatedIds, checkPathValidity } = useDataUtils();

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

    store.update(mock);

    const result = await _fetch<DataSchema | undefined, DraftData>({
      ctx: `${CRUD.CREATE} ${data.type.toUpperCase()}`,
      url: 'data/create',
      method: HTTPMethod.POST,
      body: data,
    })
      .onSuccess((res) => {
        if (!res.data) throw new Error('No data returned from create API');
        store.updateById(mock.id, res.data);
      })
      .onError((_err) => {
        store.deleteById(mock.id);
      });

    if (!result.ok || !result.response.data) return;

    return result.response.data;
  }

  async function handleRead(): Promise<DataSchema[] | undefined> {
    const result = await _fetch<DataSchema[]>({
      ctx: CRUD.READ,
      url: 'data/read',
      method: HTTPMethod.GET,
    });

    if (!result.ok) return;

    store.read(result.response.data);
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
      logApi.warn(CRUD.UPDATE, `No item found with id ${id}`);
      return;
    }

    const optimisticUpdate: Partial<DataSchema> = {
      ...data,
      updatedAt: new Date(),
    };

    store.updateById(id, optimisticUpdate);

    const result = await _fetch<DataSchema, Partial<DraftData>>({
      ctx: `${CRUD.UPDATE} ${id}`,
      url: `data/update/${id}`,
      method: HTTPMethod.PATCH,
      body: data,
    })
      .onSuccess((res) => {
        if (!res.data) throw new Error('No data returned from update API');
        store.updateById(id, res.data);
      })
      .onError(() => {
        store.updateById(id, snapshot);
      });

    if (!result.ok) return;

    await nextTick();

    return result.response.data;
  }

  async function handleDelete(): Promise<boolean> {
    const snapshot = store.getSnapshot();

    store.delete();

    const result = await _fetch<boolean>({
      ctx: CRUD.DELETE,
      url: `data/delete`,
      method: HTTPMethod.DELETE,
    }).onError((_err) => {
      logApi.warn(CRUD.DELETE, 'Error occurred, restoring previous state');
      store.read(snapshot);
    });

    if (!result.ok) return false;

    await nextTick();

    return result.response.data;
  }

  async function handleDeleteId(id: string): Promise<boolean> {
    const snapshot = store.getSnapshot(id);

    if (!snapshot) {
      logApi.warn(CRUD.DELETE, `No item found with id ${id}`);
      return false;
    }

    const relatedIds = getRelatedIds(id, { includeSelf: true });

    const batchedId: string[][] = [];
    let currentBatch: string[] = [];
    let currentLength = 0;
    const maxLength = 20;

    for (const id of relatedIds) {
      // +1 accounts for the comma separator when joining ids.
      const nextLength =
        currentLength + id.length + (currentBatch.length ? 1 : 0);

      if (currentBatch.length && nextLength > maxLength) {
        batchedId.push(currentBatch);
        currentBatch = [];
        currentLength = 0;
      }

      currentBatch.push(id);
      currentLength += id.length + (currentBatch.length > 1 ? 1 : 0);
    }

    if (currentBatch.length) batchedId.push(currentBatch);

    console.debug(
      `[DataStore] Deleting ids in ${batchedId.length} batches:`,
      batchedId
    );

    await Promise.all(batchedId.map((batch) => store.deleteById(batch)));

    const result = await _fetch<boolean>({
      ctx: `${CRUD.DELETE} ${id}`,
      url: `data/delete/${relatedIds.join(',')}`,
      method: HTTPMethod.DELETE,
    }).onError((_err) => {
      store.readById(snapshot);
    });

    if (!result.ok) return false;

    await nextTick();

    return result.response.data;
  }

  // CRUD Advanced Operations
  async function handleCreateFolder({
    path = [],
    title = 'New Folder', //i18n
    childrenIds = [],
  }: Partial<CreateFolder> = {}): Promise<DataSchema | undefined> {
    return await handleCreate({
      path,
      type: ItemType.FOLDER,
      title,
      childrenIds,
    });
  }

  async function handleCreateFolderInFolder({
    folder,
    title,
  }: CreateFolderInFolderParameters) {
    const result = await handleCreateFolder({
      path: folder.path,
      title,
      childrenIds: [],
    });
    if (!result) return;

    const children = new Set([...folder.childrenIds, result.id]);

    await handleUpdate(folder.id, {
      childrenIds: Array.from(children),
    });

    return result;
  }

  async function handleCreateNote({
    path = [],
    content = '',
  }: Partial<CreateNote> = {}): Promise<DataSchema | undefined> {
    return await handleCreate({
      path,
      type: ItemType.NOTE,
      content,
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
    create: handleCreate,
    createFolder: handleCreateFolder,
    createFolderInFolder: handleCreateFolderInFolder,
    createNote: handleCreateNote,
    createNoteInFolder: handleCreateNoteInFolder,
    read: handleRead,
    update: handleUpdate,
    delete: handleDelete,
    deleteById: handleDeleteId,
  };
}
