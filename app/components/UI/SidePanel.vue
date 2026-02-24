<script setup lang="ts">
import { useElementBounding, useEventListener, useMouse } from '@vueuse/core';

defineOptions({ inheritAttrs: false });

export interface SidePanelProps {
  minWidth: number;
  maxWidth: number;
  preferedWidth: number | null;
  collapsedWidth: number;
  collapsible: boolean;
  resizeable: boolean;
  side: 'left' | 'right';
}

const props = withDefaults(defineProps<Partial<SidePanelProps>>(), {
  minWidth: 200,
  maxWidth: 500,
  collapsedWidth: 60,
  preferedWidth: 250,
  collapsible: true,
  resizeable: true,
  side: 'left',
});

const panelRef = useTemplateRef<HTMLDivElement>('panelRef');
const width = ref<number | null>(props.preferedWidth);
const isResizing = ref(false);
const isCollapsed = defineModel<boolean>('collapsed', { default: false });

const { left, right } = useElementBounding(panelRef);
const { x: mouseX } = useMouse();

const panelStyle = computed(() => {
  const _width =
    isCollapsed.value && props.collapsible ? props.collapsedWidth : width.value;

  return { width: `${_width}px` };
});

function handleResizeMouseDown(event: MouseEvent) {
  const target = event.target as HTMLElement;
  const rect = target.getBoundingClientRect();
  const isOnResizeEdge =
    props.side === 'left'
      ? event.clientX >= rect.right - 8
      : event.clientX <= rect.left + 8;

  if (isOnResizeEdge) {
    isResizing.value = true;
  }
}

function handleMouseMove() {
  if (!isResizing.value) return;

  const newWidth =
    props.side === 'left'
      ? mouseX.value - left.value
      : right.value - mouseX.value;

  width.value = props.resizeable
    ? Math.max(props.minWidth, Math.min(props.maxWidth, newWidth))
    : props.preferedWidth;
}

function handleMouseUp() {
  isResizing.value = false;
}

useEventListener(document, 'mousemove', handleMouseMove);
useEventListener(document, 'mouseup', handleMouseUp);
</script>

<template>
  <div
    ref="panelRef"
    v-bind="$attrs"
    class="bg-default relative h-full w-fit shrink-0 overflow-hidden rounded-2xl transition-[width] duration-200 ease-in-out"
    :data-state="isCollapsed ? 'collapsed' : 'expanded'"
    :data-collapsible="collapsible"
    :data-resizeable="resizeable"
    :data-side="side"
    :data-resizing="isResizing"
    :style="panelStyle"
    @mousedown="resizeable && handleResizeMouseDown($event)"
  >
    <div v-if="isCollapsed" class="flex h-full flex-col items-center py-3">
      <slot name="collapsed" />
    </div>
    <div v-else class="h-full p-3">
      <slot />
    </div>

    <!-- Resize handle -->
    <div
      v-if="resizeable"
      class="group/handle absolute top-0 flex h-full w-2 cursor-col-resize items-center"
      :class="side === 'left' ? 'right-0' : 'left-0'"
    >
      <div
        class="mx-auto h-8 w-0.5 rounded-full transition-colors duration-150"
        :class="
          isResizing
            ? 'bg-primary'
            : 'bg-transparent group-hover/handle:bg-accented'
        "
      />
    </div>
  </div>
</template>
