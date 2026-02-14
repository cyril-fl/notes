import type { CreateFolder, CreateNote } from '~/types/data';

interface GetRelatedOptions {
  includeSelf: boolean;
}

interface CheckPathValidityOptions {
  throwError: boolean;
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
    options: { types: T }
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
    checkPathValidity,
  };
}

export function createDataApiRegistrar() {
  return () => {
    const { $hooks } = useNuxtApp();

    const store = useDataStore();
    const { fetch: _fetch } = useDataApi();
    const { getById, getRelatedIds, checkPathValidity } = useDataUtils();

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

      $hooks.callHook('on:update', mock);

      const result = await _fetch<DataSchema | undefined, DraftData>({
        ctx: `${CRUD.CREATE} ${data.type.toUpperCase()}`,
        url: 'data/create',
        method: HTTPMethod.POST,
        body: data,
      })
        .onSuccess((res) => {
          if (!res.data) throw new Error('No data returned from create API');
          $hooks.callHook('on:update:id', mock.id, res.data);
        })
        .onError((_err) => {
          $hooks.callHook('on:delete:id', mock.id);
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

      $hooks.callHook('on:read', result.response.data);
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

      $hooks.callHook('on:update:id', id, optimisticUpdate);

      const result = await _fetch<DataSchema, Partial<DraftData>>({
        ctx: `${CRUD.UPDATE} ${id}`,
        url: `data/update/${id}`,
        method: HTTPMethod.PATCH,
        body: data,
      })
        .onSuccess((res) => {
          if (!res.data) throw new Error('No data returned from update API');
          $hooks.callHook('on:update:id', id, res.data);
        })
        .onError(() => {
          $hooks.callHook('on:update:id', id, snapshot);
        });

      if (!result.ok) return;

      await nextTick();

      return result.response.data;
    }

    async function handleDelete(): Promise<boolean> {
      const snapshot = store.getSnapshot();

      $hooks.callHook('on:delete');

      const result = await _fetch<boolean>({
        ctx: CRUD.DELETE,
        url: `data/delete`,
        method: HTTPMethod.DELETE,
      }).onError((_err) => {
        logApi.warn(CRUD.DELETE, 'Error occurred, restoring previous state');
        $hooks.callHook('on:read', snapshot);
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

      $hooks.callHook('on:delete:id', relatedIds);

      const result = await _fetch<boolean>({
        ctx: `${CRUD.DELETE} ${id}`,
        url: `data/delete/${relatedIds.join(',')}`,
        method: HTTPMethod.DELETE,
      }).onError((_err) => {
        $hooks.callHook('on:read:id', snapshot);
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

    $hooks.addHooks({
      // C
      'data:create': handleCreate,
      'data:create:folder': handleCreateFolder,
      'data:create:folder:in-folder': handleCreateFolderInFolder,
      'data:create:note': handleCreateNote,
      'data:create:note:in-folder': handleCreateNoteInFolder,
      // R
      'data:read': handleRead,
      // U
      'data:update': handleUpdate,
      // D
      'data:delete': handleDelete,
      'data:delete:id': handleDeleteId,
    });
  };
}
