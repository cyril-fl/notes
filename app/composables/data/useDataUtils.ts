import type { FolderPublic, NotePublic } from '~~/shared/types/models';

interface Options {
  types: ItemType;
}

export function useDataUtils() {
  const store = useDataStore();
  const { map } = storeToRefs(store);

  function getById(
    id: string,
    options?: Partial<Omit<Options, 'types'>> & { types: ItemType.FOLDER }
  ): FolderPublic | undefined;
  function getById(
    id: string,
    options?: Partial<Omit<Options, 'types'>> & { types: ItemType.NOTE }
  ): NotePublic | undefined;
  function getById(
    id: string,
    options?: Partial<Options>
  ): FolderPublic | NotePublic | undefined {
    const result = map.value.get(id);

    if (!result) return;
    const desiredTypes = options?.types;

    const isTypeAsked = !!desiredTypes;
    const isTypeNotMatching = result.type !== desiredTypes;

    return !(isTypeAsked && isTypeNotMatching) ? result : undefined;
  }

  return {
    getById,
  };
}
