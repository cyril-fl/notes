import { draggable, dropTargetForElements } from '@atlaskit/pragmatic-drag-and-drop/element/adapter';
import { unrefElement, type MaybeElementRef } from '@vueuse/core';
import type { Item } from '~/types/items';

// A template ref: an element, or a component such as NuxtLink.
type Target = Readonly<Ref<unknown>>;

function elementOf(target: Target) {
  return unrefElement(target as MaybeElementRef) as HTMLElement | null | undefined;
}

/** Lets an element be dragged as this item, with the rest of the selection when it is part of it. */
export function useDragItem(target: Target, item: () => Item) {
  const selection = useSelection(); // Here, in setup: useRoute does not work during a drag.
  watchEffect((onCleanup) => {
    const element = elementOf(target);
    if (!element) return;
    onCleanup(
      draggable({
        element,
        getInitialData: () => {
          const ids = selection.value.includes(item().id) ? selection.value : [item().id];
          return { itemId: item().id, itemIds: ids, parentId: item().parent_id ?? null };
        },
      })
    );
  });
}

/** Moves what is dropped on an element into this folder, null being the root. */
export function useDropFolder(target: Target, folderId: () => string | null) {
  const { moveItem } = useItems();
  const isOver = ref(false);

  watchEffect((onCleanup) => {
    const element = elementOf(target);
    if (!element) return;
    onCleanup(
      dropTargetForElements({
        element,
        // contains: a folder of the tree cannot go into itself or a subfolder rendered inside it.
        canDrop: ({ source }) =>
          typeof source.data.itemId === 'string' &&
          source.data.parentId !== folderId() &&
          source.data.itemId !== folderId() &&
          !(source.data.itemIds as string[]).includes(folderId() ?? '') &&
          !source.element.contains(element),
        // Targets nest (a folder row inside the panel of its parent): only the innermost one reacts.
        onDropTargetChange: ({ location, self }) => {
          isOver.value = location.current.dropTargets[0]?.element === self.element;
        },
        onDragLeave: () => {
          isOver.value = false;
        },
        onDrop: ({ source, location, self }) => {
          isOver.value = false;
          if (location.current.dropTargets[0]?.element !== self.element) return;
          moveItem(source.data.itemIds as string[], folderId());
        },
      })
    );
  });

  return isOver;
}
