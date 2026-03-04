import { draggable } from '@atlaskit/pragmatic-drag-and-drop/element/adapter';
import { setCustomNativeDragPreview } from '@atlaskit/pragmatic-drag-and-drop/element/set-custom-native-drag-preview';
import { createApp, type Component } from 'vue';

export interface DragData {
  type: 'items';
  ids: string[];
}

interface UseDraggableOptions {
  elementRef: Ref<HTMLElement | null>;
  item: Ref<Data>;
  ghostComponent: Component;
  canDragCheck?: () => boolean;
}

export function useDraggable({
  elementRef,
  item,
  ghostComponent,
  canDragCheck,
}: UseDraggableOptions) {
  const { selectedIds, isSelected, clearSelection } = useSelection();
  const { getById } = useDataUtils();
  const isDragging = ref(false);

  let cleanup: (() => void) | null = null;

  function setup() {
    const el = elementRef.value;
    if (!el) return;

    cleanup?.();

    cleanup = draggable({
      element: el,
      getInitialData(): DragData {
        // If the item is selected, drag all selected items
        if (isSelected(item.value.id)) {
          return { type: 'items', ids: Array.from(selectedIds.value) };
        }
        // Otherwise, clear selection and drag just this item
        clearSelection();
        return { type: 'items', ids: [item.value.id] };
      },
      canDrag() {
        if (!item.value) return false;
        if (item.value.isDeleted || item.value.id === 'root') return false;
        if (canDragCheck && !canDragCheck()) return false;
        return true;
      },
      onGenerateDragPreview({ nativeSetDragImage }) {
        setCustomNativeDragPreview({
          nativeSetDragImage,
          render({ container }) {
            const dragIds = isSelected(item.value.id)
              ? Array.from(selectedIds.value)
              : [item.value.id];

            const items = dragIds
              .map((id) => getById(id))
              .filter((i): i is Data => !!i);

            const app = createApp(ghostComponent, { items });
            app.mount(container);

            return () => app.unmount();
          },
        });
      },
      onDragStart() {
        isDragging.value = true;
      },
      onDrop() {
        isDragging.value = false;
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

  return { isDragging };
}
