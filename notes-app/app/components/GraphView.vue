<script setup lang="ts">
import type { NotesGraph } from '~/types/items';

const { icons } = useIcons();

// tag: a note to one of its tags; otherwise a [[link]] between two notes.
type Edge = { source: string; target: string; tag: boolean };

// mini: in a narrow panel, icons instead of titles, and the card of the hovered note in a popover drawn out of the panel.
const props = defineProps<{ graph: NotesGraph; selected?: string; mini?: boolean }>();

const hovered = ref<string | null>(null);
// The card keeps the last hovered note: the popover still draws it while it fades out.
const carded = ref<string | null>(null);
const anchor = ref<HTMLElement>();
const [DefineCard, ReuseCard] = createReusableTemplate();

function hover(id: string, event: Event) {
  hovered.value = id;
  carded.value = id;
  anchor.value = event.currentTarget as HTMLElement;
}
// A tag lights up its notes, without a card.
function hoverTag(tag: string) {
  hovered.value = tagNode(tag);
}
const active = computed(() => hovered.value ?? props.selected ?? null);

const positions = computed(() => layoutGraph(props.graph.notes, props.graph.links));

// The tags are nodes too, each with how many notes carry it.
const tags = computed(() => {
  const counts = new Map<string, number>();
  for (const note of props.graph.notes) for (const tag of note.tags ?? []) counts.set(tag, (counts.get(tag) ?? 0) + 1);
  return counts;
});
const edges = computed<Edge[]>(() => [
  ...props.graph.links.map((link) => ({ ...link, tag: false })),
  ...props.graph.notes.flatMap((note) =>
    (note.tags ?? []).map((tag) => ({ source: note.id, target: tagNode(tag), tag: true })),
  ),
]);

function isActiveLink(edge: Edge) {
  return edge.source === active.value || edge.target === active.value;
}

const activeLinks = computed(() => edges.value.filter(isActiveLink));

const linkedToActive = computed(() => new Set(activeLinks.value.flatMap((link) => [link.source, link.target])));

const hoveredNote = computed(() => props.graph.notes.find((note) => note.id === hovered.value) ?? null);
const cardNote = computed(() => props.graph.notes.find((note) => note.id === carded.value) ?? null);
const cardLinks = computed(() =>
  props.graph.links.filter((link) => link.source === carded.value || link.target === carded.value),
);

const titles = computed(() => new Map(props.graph.notes.map((note) => [note.id, note.title])));

function otherEnd(link: Edge | NotesGraph['links'][number]) {
  return link.source === carded.value ? link.target : link.source;
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
// mini: the bare wheel scrolls the panel, only ⌘/Ctrl + wheel or a pinch zooms.
useEventListener(
  view,
  'wheel',
  (event: WheelEvent) => {
    if (props.mini && !event.ctrlKey && !event.metaKey) return;
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

function iconClass(id: string) {
  if (id === active.value) return 'text-primary';
  return isDimmed(id) ? 'text-(--ui-border-accented)' : 'text-muted';
}

function tagClass(tag: string) {
  if (tagNode(tag) === active.value) return 'font-medium text-primary';
  return isDimmed(tagNode(tag)) ? 'text-dimmed opacity-30' : 'text-muted';
}

function linkClass(link: Edge) {
  if (active.value === null) return 'stroke-(--ui-border-accented)';
  return isActiveLink(link) ? 'stroke-(--ui-primary)' : 'stroke-(--ui-border-muted)';
}
</script>

<template>
  <div
    ref="view"
    class="relative size-full cursor-grab touch-none overflow-hidden select-none active:cursor-grabbing"
    :class="{ 'rounded-md bg-muted': mini }"
    @pointerdown="startDrag"
    @click.capture="swallowClickAfterDrag"
  >
    <div class="absolute inset-0" :style="{ transform: `translate(${offset.x}px, ${offset.y}px)` }">
      <svg class="absolute inset-0 size-full" viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true">
        <line
          v-for="edge in edges"
          :key="`${edge.source}-${edge.target}`"
          :x1="at(edge.source).x"
          :y1="at(edge.source).y"
          :x2="at(edge.target).x"
          :y2="at(edge.target).y"
          :stroke-width="edge.tag ? 1 : 2"
          :stroke-dasharray="edge.tag ? '0.5 0.5' : undefined"
          :class="linkClass(edge)"
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
        :class="[{ 'z-10': note.id === active }, mini ? '-translate-y-1/2' : '-translate-y-[5px]']"
        :aria-label="mini ? displayTitle(note.title, $t('untitled')) : undefined"
        class="absolute flex max-w-32 -translate-x-1/2 flex-col items-center gap-1 rounded-md"
        @mouseenter="hover(note.id, $event)"
        @mouseleave="hovered = null"
        @focus="hover(note.id, $event)"
        @blur="hovered = null"
      >
        <!-- On the view's background: the links drawn under the icon must not show through it. -->
        <span v-if="mini" class="flex rounded-sm bg-muted p-0.5">
          <UIcon
            :name="icons.note"
            class="size-4 shrink-0 transition-colors duration-150"
            :class="iconClass(note.id)"
          />
        </span>
        <template v-else>
          <!-- Solid colors, not opacity: links drawn under the dot must not show through it. -->
          <span class="size-2.5 shrink-0 rounded-full transition-colors duration-150" :class="dotClass(note.id)" />
          <span
            class="max-w-full truncate text-xs transition-opacity duration-150"
            :class="{ 'font-medium': note.id === active, 'opacity-30': isDimmed(note.id) }"
            >{{ displayTitle(note.title, $t('untitled')) }}</span
          >
        </template>
      </NuxtLink>

      <NuxtLink
        v-for="[tag] in tags"
        :key="tagNode(tag)"
        :to="tagPath(tag)"
        draggable="false"
        :style="{ left: `${at(tagNode(tag)).x}%`, top: `${at(tagNode(tag)).y}%` }"
        :class="{ 'z-10': tagNode(tag) === active }"
        :aria-label="mini ? `#${tag}` : undefined"
        class="absolute max-w-32 -translate-x-1/2 -translate-y-1/2 rounded-sm px-0.5"
        @mouseenter="hoverTag(tag)"
        @mouseleave="hovered = null"
        @focus="hoverTag(tag)"
        @blur="hovered = null"
      >
        <!-- On the view's background: the links drawn under the tag must not show through it. -->
        <span v-if="mini" class="flex rounded-sm bg-muted p-0.5">
          <UIcon :name="icons.tag" class="size-3.5 shrink-0 transition-colors duration-150" :class="tagClass(tag)" />
        </span>
        <span v-else class="block truncate bg-default text-xs transition-colors duration-150" :class="tagClass(tag)"
          >#{{ tag }}</span
        >
      </NuxtLink>

      <DefineCard v-if="cardNote">
        <span class="shrink-0 truncate font-medium">{{ displayTitle(cardNote.title, $t('untitled')) }}</span>
        <span class="shrink-0 truncate text-muted">{{ cardNote.path }}</span>
        <template v-if="cardLinks.length || cardNote.tags?.length">
          <span v-if="cardNote.tags?.length" class="truncate text-primary">{{
            cardNote.tags.map((tag) => `#${tag}`).join(' ')
          }}</span>
          <ul class="mt-1 flex min-h-0 flex-col gap-0.5 overflow-hidden border-t border-default pt-1 empty:hidden">
            <li v-for="link in cardLinks" :key="`${link.source}-${link.target}`" class="flex gap-2">
              <span class="truncate">{{ displayTitle(titles.get(otherEnd(link)), $t('untitled')) }}</span>
              <span class="ml-auto shrink-0 text-primary">{{ $t('graph.link') }}</span>
            </li>
          </ul>
        </template>
        <span v-else class="text-muted">{{ $t('graph.noLink') }}</span>
      </DefineCard>
      <div
        v-if="hoveredNote && !mini"
        :style="{ left: `${at(hoveredNote.id).x}%`, top: `${at(hoveredNote.id).y}%` }"
        class="pointer-events-none absolute z-20 ml-4 flex h-40 w-64 -translate-y-1/2 flex-col gap-1 rounded-md bg-default p-2 text-xs shadow-lg ring ring-default"
      >
        <ReuseCard />
      </div>
    </div>

    <UPopover
      v-if="mini"
      mode="hover"
      :open="!!hoveredNote"
      :reference="anchor"
      :content="{ side: 'left' }"
      :ui="{ content: 'pointer-events-none flex max-h-40 w-64 flex-col gap-1 p-2 text-xs' }"
    >
      <template #content><ReuseCard /></template>
    </UPopover>

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
      <UTooltip v-if="mini && selected" :text="$t('related.showGraph')">
        <UButton
          :icon="icons.expand"
          :to="{ path: '/graph', query: { note: selected } }"
          :aria-label="$t('related.showGraph')"
          color="neutral"
          variant="ghost"
          size="sm"
        />
      </UTooltip>
    </div>
  </div>
</template>
