<script setup lang="ts">
import type { NotesGraph } from '~/types/items';

type Link = NotesGraph['links'][number];

const props = defineProps<{ graph: NotesGraph; selected?: string }>();

const hovered = ref<string | null>(null);
const active = computed(() => hovered.value ?? props.selected ?? null);

const positions = computed(() => layoutGraph(props.graph.notes, props.graph.links));

function isActiveLink(link: Link) {
  return link.source === active.value || link.target === active.value;
}

const activeLinks = computed(() => props.graph.links.filter(isActiveLink));

const linkedToActive = computed(() => new Set(activeLinks.value.flatMap(link => [link.source, link.target])));

function isDimmed(id: string) {
  return active.value !== null && id !== active.value && !linkedToActive.value.has(id);
}

function at(id: string) {
  return positions.value.get(id) ?? { x: 50, y: 50 };
}

function linkLabelStyle(link: Link) {
  const from = at(link.source);
  const to = at(link.target);
  return { left: `${(from.x + to.x) / 2}%`, top: `${(from.y + to.y) / 2}%` };
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

// A drag that ends on a note must not open it.
function swallowClickAfterDrag(event: MouseEvent) {
  if (!moved) return;
  event.preventDefault();
  event.stopPropagation();
  moved = false;
}

function linkClass(link: Link) {
  if (active.value === null) return 'stroke-(--ui-border-accented)';
  return isActiveLink(link) ? 'stroke-(--ui-primary)' : 'stroke-(--ui-border-muted)';
}
</script>

<template>
  <div
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
        :title="note.path"
        draggable="false"
        :style="{ left: `${at(note.id).x}%`, top: `${at(note.id).y}%` }"
        :class="{ 'opacity-30': isDimmed(note.id), 'z-10': note.id === active }"
        class="absolute flex max-w-32 -translate-x-1/2 -translate-y-[5px] flex-col items-center gap-1 rounded-md transition-opacity duration-150"
        @mouseenter="hovered = note.id"
        @mouseleave="hovered = null"
        @focus="hovered = note.id"
        @blur="hovered = null"
      >
        <span class="size-2.5 shrink-0 rounded-full" :class="note.id === active ? 'bg-primary' : 'bg-inverted/40'" />
        <span class="max-w-full truncate text-xs" :class="{ 'font-medium': note.id === active }">{{ displayTitle(note.title) }}</span>
      </NuxtLink>

      <span
        v-for="link in activeLinks"
        :key="`${link.source}-${link.target}`"
        :style="linkLabelStyle(link)"
        class="pointer-events-none absolute z-20 -translate-x-1/2 -translate-y-1/2 rounded-sm bg-default px-1 text-xs whitespace-nowrap text-primary"
      >
        {{ [...(link.linked ? ['lien'] : []), ...link.tags.map(tag => `#${tag}`)].join(' · ') }}
      </span>
    </div>
  </div>
</template>
