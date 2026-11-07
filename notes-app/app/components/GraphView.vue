<script setup lang="ts">
import type { NotesGraph } from '~/types/items';

const { icons } = useIcons();

type Link = NotesGraph['links'][number];

const props = defineProps<{ graph: NotesGraph; selected?: string }>();

const hovered = ref<string | null>(null);
const active = computed(() => hovered.value ?? props.selected ?? null);

const positions = computed(() => layoutGraph(props.graph.notes, props.graph.links));

function isActiveLink(link: Link) {
  return link.source === active.value || link.target === active.value;
}

const activeLinks = computed(() => props.graph.links.filter(isActiveLink));

const linkedToActive = computed(() => new Set(activeLinks.value.flatMap((link) => [link.source, link.target])));

const hoveredNote = computed(() => props.graph.notes.find((note) => note.id === hovered.value) ?? null);

const titles = computed(() => new Map(props.graph.notes.map((note) => [note.id, note.title])));

function otherEnd(link: Link) {
  return link.source === hovered.value ? link.target : link.source;
}

function isDimmed(id: string) {
  return active.value !== null && id !== active.value && !linkedToActive.value.has(id);
}

// Zoom spreads the notes from the center rather than scaling the drawing: titles keep their size and stop overlapping.
const zoom = ref(1);

function at(id: string) {
  const { x, y } = positions.value.get(id) ?? { x: 50, y: 50 };
  return { x: 50 + (x - 50) * zoom.value, y: 50 + (y - 50) * zoom.value };
}

// Drag the background to move around the graph.
const offset = ref({ x: 0, y: 0 });
let drag: { x: number; y: number; from: { x: number; y: number } } | null = null;
let moved = false;

function startDrag(event: PointerEvent) {
  if (event.button !== 0) return;
  drag = { x: event.clientX, y: event.clientY, from: offset.value };
  moved = false;
}

useEventListener(window, 'pointermove', (event: PointerEvent) => {
  if (!drag) return;
  const dx = event.clientX - drag.x;
  const dy = event.clientY - drag.y;
  if (Math.hypot(dx, dy) > 3) moved = true;
  offset.value = { x: drag.from.x + dx, y: drag.from.y + dy };
});
useEventListener(window, 'pointerup', () => (drag = null));

const view = useTemplateRef('view');
const MIN_ZOOM = 0.5;
const MAX_ZOOM = 6;

/** Zooms by factor, keeping the point under (x, y) — in px from the view's corner, its center by default — in place. */
function zoomBy(factor: number, x?: number, y?: number) {
  const { width, height } = view.value!.getBoundingClientRect();
  const next = Math.min(Math.max(zoom.value * factor, MIN_ZOOM), MAX_ZOOM);
  const ratio = next / zoom.value;
  const cx = (x ?? width / 2) - width / 2;
  const cy = (y ?? height / 2) - height / 2;
  offset.value = { x: cx - (cx - offset.value.x) * ratio, y: cy - (cy - offset.value.y) * ratio };
  zoom.value = next;
}

// The wheel, and a trackpad pinch (a wheel with ctrlKey for the browser); passive: false to keep the page from scrolling.
useEventListener(
  view,
  'wheel',
  (event: WheelEvent) => {
    event.preventDefault();
    const rect = view.value!.getBoundingClientRect();
    zoomBy(
      Math.exp(-event.deltaY * (event.ctrlKey ? 0.01 : 0.002)),
      event.clientX - rect.left,
      event.clientY - rect.top,
    );
  },
  { passive: false },
);

function recenter() {
  zoom.value = 1;
  offset.value = { x: 0, y: 0 };
}

// A drag that ends on a note must not open it.
function swallowClickAfterDrag(event: MouseEvent) {
  if (!moved) return;
  event.preventDefault();
  event.stopPropagation();
  moved = false;
}

function dotClass(id: string) {
  if (id === active.value) return 'bg-primary';
  return isDimmed(id) ? 'bg-(--ui-border-accented)' : 'bg-(--ui-text-dimmed)';
}

function linkClass(link: Link) {
  if (active.value === null) return 'stroke-(--ui-border-accented)';
  return isActiveLink(link) ? 'stroke-(--ui-primary)' : 'stroke-(--ui-border-muted)';
}
</script>

<template>
  <div
    ref="view"
    class="relative size-full cursor-grab touch-none overflow-hidden select-none active:cursor-grabbing"
    @pointerdown="startDrag"
    @click.capture="swallowClickAfterDrag"
  >
    <div class="absolute inset-0" :style="{ transform: `translate(${offset.x}px, ${offset.y}px)` }">
      <svg class="absolute inset-0 size-full" viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true">
        <line
          v-for="link in graph.links"
          :key="`${link.source}-${link.target}`"
          :x1="at(link.source).x"
          :y1="at(link.source).y"
          :x2="at(link.target).x"
          :y2="at(link.target).y"
          :stroke-width="Math.min(link.tags.length + (link.linked ? 1 : 0), 4)"
          :stroke-dasharray="link.linked ? undefined : '0.5 0.5'"
          :class="linkClass(link)"
          class="transition-colors duration-150"
          vector-effect="non-scaling-stroke"
        />
      </svg>

      <NuxtLink
        v-for="note in graph.notes"
        :key="note.id"
        :to="`/notes/${note.id}`"
        draggable="false"
        :style="{ left: `${at(note.id).x}%`, top: `${at(note.id).y}%` }"
        :class="{ 'z-10': note.id === active }"
        class="absolute flex max-w-32 -translate-x-1/2 -translate-y-[5px] flex-col items-center gap-1 rounded-md"
        @mouseenter="hovered = note.id"
        @mouseleave="hovered = null"
        @focus="hovered = note.id"
        @blur="hovered = null"
      >
        <!-- Solid colors, not opacity: links drawn under the dot must not show through it. -->
        <span class="size-2.5 shrink-0 rounded-full transition-colors duration-150" :class="dotClass(note.id)" />
        <span
          class="max-w-full truncate text-xs transition-opacity duration-150"
          :class="{ 'font-medium': note.id === active, 'opacity-30': isDimmed(note.id) }"
          >{{ displayTitle(note.title, $t('untitled')) }}</span
        >
      </NuxtLink>

      <div
        v-if="hoveredNote"
        :style="{ left: `${at(hoveredNote.id).x}%`, top: `${at(hoveredNote.id).y}%` }"
        class="pointer-events-none absolute z-20 ml-4 flex h-40 w-64 -translate-y-1/2 flex-col gap-1 rounded-md bg-default p-2 text-xs shadow-lg ring ring-default"
      >
        <span class="shrink-0 truncate font-medium">{{ displayTitle(hoveredNote.title, $t('untitled')) }}</span>
        <span class="shrink-0 truncate text-muted">{{ hoveredNote.path }}</span>
        <ul
          v-if="activeLinks.length"
          class="mt-1 flex min-h-0 flex-col gap-0.5 overflow-hidden border-t border-default pt-1"
        >
          <li v-for="link in activeLinks" :key="`${link.source}-${link.target}`" class="flex gap-2">
            <span class="truncate">{{ displayTitle(titles.get(otherEnd(link)), $t('untitled')) }}</span>
            <span class="ml-auto shrink-0 text-primary">{{
              [...(link.linked ? [$t('graph.link')] : []), ...link.tags.map((tag) => `#${tag}`)].join(' ')
            }}</span>
          </li>
        </ul>
        <span v-else class="text-muted">{{ $t('graph.noLink') }}</span>
      </div>
    </div>

    <!-- @pointerdown.stop: a click on a button is not the start of a drag. -->
    <div
      class="absolute right-3 bottom-3 z-30 flex gap-1 rounded-md bg-default/80 p-0.5 backdrop-blur"
      @pointerdown.stop
    >
      <UTooltip :text="$t('graph.zoomOut')">
        <UButton
          :icon="icons.zoomOut"
          color="neutral"
          variant="ghost"
          size="sm"
          :aria-label="$t('graph.zoomOut')"
          :disabled="zoom <= MIN_ZOOM"
          @click="zoomBy(1 / 1.4)"
        />
      </UTooltip>
      <UTooltip :text="$t('graph.zoomIn')">
        <UButton
          :icon="icons.zoomIn"
          color="neutral"
          variant="ghost"
          size="sm"
          :aria-label="$t('graph.zoomIn')"
          :disabled="zoom >= MAX_ZOOM"
          @click="zoomBy(1.4)"
        />
      </UTooltip>
      <UTooltip :text="$t('graph.recenter')">
        <UButton
          :icon="icons.recenter"
          color="neutral"
          variant="ghost"
          size="sm"
          :aria-label="$t('graph.recenter')"
          @click="recenter"
        />
      </UTooltip>
    </div>
  </div>
</template>
