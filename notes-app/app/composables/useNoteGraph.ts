import type { NotesGraph } from '~/types/items';

/** The graph around one note, laid out by the server: the graph page from a note, and the mini graph of its panel. */
export function useNoteGraph(id: MaybeRefOrGetter<string | undefined>) {
  const api = useApi();
  return useLazyAsyncData(
    () => `graph:${toValue(id)}`,
    () => {
      const noteId = toValue(id);
      return noteId ? api<NotesGraph>(`/api/notes/${noteId}/graph`, { quiet: true }) : Promise.resolve(null);
    },
  );
}
