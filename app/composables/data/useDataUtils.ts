interface GetRelatedOptions {
  includeSelf: boolean;
}

interface CheckPathValidityOptions {
  throwError: boolean;
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
