import type { Neighbors } from '~/types/items';

/** The notes around one note (its tags, its [[links]], its backlinks): one request, shared by the related notes, the mini graph and the [[ menu. */
export function useNeighbors(id: MaybeRefOrGetter<string | undefined>) {
  const api = useApi();
  return useLazyAsyncData(
    () => `neighbors:${toValue(id)}`,
    () => {
      const noteId = toValue(id);
      return noteId ? api<Neighbors>(`/api/notes/${noteId}/neighbors`, { quiet: true }) : Promise.resolve(null);
    },
  );
}
