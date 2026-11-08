/** Adds the id to the selection, or takes it out when it is there. */
export function toggleSelected(selection: string[], id: string) {
  return selection.includes(id) ? selection.filter((other) => other !== id) : [...selection, id];
}

/** Adds the ids between two indexes, both included and in any order, without doubles. */
export function selectRange(selection: string[], ids: string[], from: number, to: number) {
  return [...new Set([...selection, ...ids.slice(Math.min(from, to), Math.max(from, to) + 1)])];
}

/** What a dragged item carries, set by useDragItem. */
export type DragData = Record<string | symbol, unknown>;

/** Whether dragged items can go into a folder (null the root): not where they already are, not into one of them. */
export function canDropInto(data: DragData, folderId: string | null) {
  return (
    typeof data.itemId === 'string' &&
    data.parentId !== folderId &&
    data.itemId !== folderId &&
    !(data.itemIds as string[]).includes(folderId ?? '')
  );
}
