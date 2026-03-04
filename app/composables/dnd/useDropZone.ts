import { dropTargetForElements } from '@atlaskit/pragmatic-drag-and-drop/element/adapter';
import type { DragData } from './useDraggable';

interface UseDropZoneOptions {
  elementRef: Ref<HTMLElement | null>;
  folderId: Ref<string | null>;
}

export function useDropZone({ elementRef, folderId }: UseDropZoneOptions) {
  const { moveItems } = useDataActions();
  const { getRelatedIds, getById } = useDataUtils();
  const isOver = ref(false);

  let cleanup: (() => void) | null = null;

  function isDragData(data: Record<string, unknown>): data is DragData {
    return data.type === 'items' && Array.isArray(data.ids);
  }

  function canAcceptDrop(data: Record<string, unknown>): boolean {
    if (!isDragData(data)) return false;

    const targetId = folderId.value ?? 'root';

    // Check if at least one item can be moved to the target
    let hasValidItem = false;

    for (const id of data.ids) {
      // Cannot drop on itself
      if (id === targetId) return false;
      // Root is not movable
      if (id === 'root') continue;

      const item = getById(id);
      if (!item) continue;

      // Already in target folder (no-op)
      if (item.ancestor === targetId) continue;

      // Cannot drop a folder into its own descendant (circular)
      if (item.type === ItemType.FOLDER) {
        const descendants = getRelatedIds(id, { includeSelf: true });
        if (descendants.includes(targetId)) return false;
      }

      hasValidItem = true;
    }

    return hasValidItem;
  }

  function setup() {
    const el = elementRef.value;
    if (!el) return;

    cleanup?.();

    cleanup = dropTargetForElements({
      element: el,
      canDrop({ source }) {
        return canAcceptDrop(source.data);
      },
      onDragEnter() {
        isOver.value = true;
      },
      onDragLeave() {
        isOver.value = false;
      },
      onDrop({ source }) {
        isOver.value = false;
        if (!isDragData(source.data)) return;
        moveItems(source.data.ids, folderId.value);
      },
    });
  }

  watch(elementRef, (el) => {
    if (el) setup();
  });

  onMounted(() => {
    if (elementRef.value) setup();
  });

  onBeforeUnmount(() => {
    cleanup?.();
  });

  return { isOver };
}
