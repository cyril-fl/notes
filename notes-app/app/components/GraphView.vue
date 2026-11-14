<script setup lang="ts">
import type { NotesGraph } from '~/types/items';

const { icons } = useIcons();
const appConfig = useAppConfig();

// tag: a note to one of its tags; otherwise a [[link]] between two notes.
type Edge = { source: string; target: string; strength: number; tag: boolean };

// mini: in a narrow panel, icons instead of titles, and the card of the hovered note in a popover drawn out of the panel.
// cache: where the settled layout is kept for the next visit (layoutCache), for a graph worth it.
const props = defineProps<{ graph: NotesGraph; selected?: string; mini?: boolean; cache?: string }>();

const hovered = ref<string | null>(null);
// The card keeps the last hovered note: the popover still draws it while it fades out.
const carded = ref<string | null>(null);
const anchor = ref<HTMLElement>();
const [DefineCard, ReuseCard] = createReusableTemplate();

function hover(id: string, event?: Event) {
  hovered.value = id;
  if (isTag(id)) return; // A tag lights up its notes, without a card.
  carded.value = id;
  if (event) anchor.value = event.currentTarget as HTMLElement;
}
const active = computed(() => hovered.value ?? props.selected ?? null);

// The layout settles in a worker, drawn once settled: a big graph takes seconds.
const positions = shallowRef(new Map<string, { x: number; y: number }>());
const settled = ref(false);
watch(
  () => props.graph,
  (graph, _, onCleanup) => {
    settled.value = false;
    const worker = new Worker(new URL('../workers/layout.ts', import.meta.url), { type: 'module' });
    worker.onmessage = ({ data }: MessageEvent<typeof positions.value>) => {
      positions.value = data;
      settled.value = true;
    };
    worker.postMessage({
      notes: graph.notes.map(({ id, tags }) => ({ id, tags: tags && [...tags] })),
      links: graph.links.map(({ source, target }) => ({ source, target })),
      cache: props.cache,
    });
    onCleanup(() => worker.terminate());
  },
  { immediate: true },
);

// The tags are nodes too, each with how many notes carry it.
const tags = computed(() => {
  const counts = new Map<string, number>();
  for (const note of props.graph.notes) for (const tag of note.tags ?? []) counts.set(tag, (counts.get(tag) ?? 0) + 1);
  return counts;
});
const isTag = (id: string) => id.startsWith('#');
// The springs of the layout, drawn: a link pulling little is drawn faint.
const edges = computed<Edge[]>(() =>
  graphEdges(props.graph.notes, props.graph.links).map(({ source, target, strength }) => ({
    source,
    target,
    strength,
    tag: isTag(target),
  })),
);
const neighbors = computed(() => {
  const ends = new Map<string, string[]>();
  for (const { source, target } of edges.value) {
    ends.set(source, [...(ends.get(source) ?? []), target]);
    ends.set(target, [...(ends.get(target) ?? []), source]);
  }
  return ends;
});

// In front: the active node, its neighbors and the links between them.
const focus = computed(() => new Set(active.value ? [active.value, ...(neighbors.value.get(active.value) ?? [])] : []));
const inFront = (edge: Edge) => focus.value.has(edge.source) && focus.value.has(edge.target);
const isDimmed = (id: string) => active.value !== null && !focus.value.has(id);

const hoveredNote = computed(() => props.graph.notes.find((note) => note.id === hovered.value) ?? null);
const cardNote = computed(() => props.graph.notes.find((note) => note.id === carded.value) ?? null);
const cardLinks = computed(() =>
  props.graph.links.filter((link) => link.source === carded.value || link.target === carded.value),
);

const titles = computed(() => new Map(props.graph.notes.map((note) => [note.id, note.title])));

function otherEnd(link: NotesGraph['links'][number]) {
  return link.source === carded.value ? link.target : link.source;
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
const TITLE_LIMIT = 200;
// Deep enough for a big graph to reach the titles: few enough nodes on screen at the deepest.
const maxZoom = computed(() => Math.max(6, 3 * Math.sqrt(positions.value.size / TITLE_LIMIT)));

/** Zooms by factor, keeping the point under (x, y) — in px from the view's corner, its center by default — in place. */
function zoomBy(factor: number, x?: number, y?: number) {
  const { width, height } = view.value!.getBoundingClientRect();
  const next = Math.min(Math.max(zoom.value * factor, MIN_ZOOM), maxZoom.value);
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

// What is on screen, in px from the view's corner; a margin keeps a dot or a title cut by the edge.
const { width, height } = useElementSize(view);
const CULL_MARGIN = 40;
function px(id: string) {
  const { x, y } = at(id);
  return { x: (x / 100) * width.value + offset.value.x, y: (y / 100) * height.value + offset.value.y };
}
const visible = computed(() => {
  const ids = new Set<string>();
  for (const id of positions.value.keys()) {
    const { x, y } = px(id);
    if (x > -CULL_MARGIN && x < width.value + CULL_MARGIN && y > -CULL_MARGIN && y < height.value + CULL_MARGIN)
      ids.add(id);
  }
  return ids;
});

// The titles, by zoom: every one once few nodes are on screen, else the biggest tags alone. The nodes in front always,
// then the tags by size, then the notes; a title landing on one already placed is left out, its node stays a dot.
// Titled nodes are links of the page (keyboard, screen readers, the e2e); the others are dots on the canvas.
const BIG_TAGS = 20;
// A title's box (max-w-32, its dot above it): two titles closer than it on both axes overlap.
const TITLE_CELL = { width: 128, height: 28 };
const titled = computed(() => {
  const bySize = (a: string, b: string) => (tags.value.get(b.slice(1)) ?? 0) - (tags.value.get(a.slice(1)) ?? 0);
  const ids = [...visible.value];
  const candidates = ids.length <= TITLE_LIMIT ? ids.sort(bySize) : ids.filter(isTag).sort(bySize).slice(0, BIG_TAGS);
  const shown = new Set<string>();
  const taken = new Map<string, { x: number; y: number }[]>();
  const place = (id: string, force: boolean) => {
    const { x, y } = px(id);
    const [cx, cy] = [Math.floor(x / TITLE_CELL.width), Math.floor(y / TITLE_CELL.height)];
    // Any title closer than a box is in this cell or one around it: free, it is placed.
    const near = [-1, 0, 1].flatMap((dx) => [-1, 0, 1].map((dy) => taken.get(`${cx + dx},${cy + dy}`) ?? []));
    if (
      !force &&
      near.flat().some((other) => Math.abs(other.x - x) < TITLE_CELL.width && Math.abs(other.y - y) < TITLE_CELL.height)
    )
      return;
    taken.set(`${cx},${cy}`, [...(taken.get(`${cx},${cy}`) ?? []), { x, y }]);
    shown.add(id);
  };
  for (const id of focus.value) if (visible.value.has(id)) place(id, true);
  // A small graph keeps every title: the collisions of its layout already keep them apart.
  for (const id of candidates) if (!shown.has(id)) place(id, positions.value.size <= TITLE_LIMIT);
  return shown;
});
const titledNotes = computed(() => props.graph.notes.filter((note) => titled.value.has(note.id)));
const titledTags = computed(() => [...tags.value.keys()].filter((tag) => titled.value.has(tagNode(tag))));

// The canvas draws the links and the untitled dots: the dimmed ones first, then the ones in front.
const canvas = useTemplateRef('canvas');
const NOTE_RADIUS = 5;
// Many nodes on screen: smaller dots and fainter links, or they melt into one stain.
const density = computed(() => Math.min(Math.max(Math.sqrt(TITLE_LIMIT / Math.max(visible.value.size, 1)), 0.3), 1));
const radius = (id: string) =>
  density.value * (isTag(id) ? 3 + Math.min(Math.sqrt(tags.value.get(id.slice(1)) ?? 1), 6) : NOTE_RADIUS);
// A theme or mode switch changes the tokens the canvas reads.
const theme = ref(0);
useMutationObserver(document.documentElement, () => theme.value++, { attributes: true });

let frame = 0;
watchEffect(() => {
  void [positions.value, visible.value, titled.value, focus.value, active.value, theme.value];
  cancelAnimationFrame(frame);
  frame = requestAnimationFrame(draw);
});
onBeforeUnmount(() => cancelAnimationFrame(frame));

function draw() {
  const el = canvas.value;
  if (!el) return;
  const ratio = window.devicePixelRatio;
  el.width = width.value * ratio;
  el.height = height.value * ratio;
  const ctx = el.getContext('2d')!;
  ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
  const style = getComputedStyle(el);
  const color = (token: string) => style.getPropertyValue(token).trim();

  const back: Edge[] = [];
  const front: Edge[] = [];
  for (const edge of edges.value)
    if (visible.value.has(edge.source) || visible.value.has(edge.target)) (inFront(edge) ? front : back).push(edge);
  const untitled = [...visible.value].filter((id) => !titled.value.has(id));

  strokeEdges(ctx, back, color(active.value ? '--ui-border-muted' : '--ui-border-accented'), density.value);
  fillDots(
    ctx,
    untitled.filter((id) => !focus.value.has(id)),
    color(active.value ? '--ui-border-accented' : '--ui-text-dimmed'),
  );
  strokeEdges(
    ctx,
    front.map((edge) => ({ ...edge, strength: 1 })), // In front, every link at full strength.
    color('--ui-primary'),
    1,
  );
  fillDots(
    ctx,
    untitled.filter((id) => focus.value.has(id)),
    color('--ui-text-dimmed'),
  );
}

// A [[link]] in a plain line, a note to its tag dashed; by how hard it pulls, in a few steps of opacity: one path each.
const OPACITY_STEPS = 4;
function strokeEdges(ctx: CanvasRenderingContext2D, list: Edge[], stroke: string, opacity: number) {
  ctx.strokeStyle = stroke;
  for (const tag of [false, true])
    for (let step = 1; step <= OPACITY_STEPS; step++) {
      ctx.lineWidth = tag ? 1 : 2;
      ctx.setLineDash(tag ? [4, 4] : []);
      ctx.globalAlpha = opacity * (step / OPACITY_STEPS);
      ctx.beginPath();
      for (const edge of list) {
        if (edge.tag !== tag || Math.max(Math.ceil(Math.min(edge.strength, 1) * OPACITY_STEPS), 1) !== step) continue;
        const from = px(edge.source);
        const to = px(edge.target);
        ctx.moveTo(from.x, from.y);
        ctx.lineTo(to.x, to.y);
      }
      ctx.stroke();
    }
  ctx.globalAlpha = 1;
}

function fillDots(ctx: CanvasRenderingContext2D, ids: string[], fill: string) {
  ctx.fillStyle = fill;
  ctx.beginPath();
  for (const id of ids) {
    const { x, y } = px(id);
    const r = radius(id);
    ctx.moveTo(x + r, y);
    ctx.arc(x, y, r, 0, 2 * Math.PI);
  }
  ctx.fill();
}

// An untitled dot under the pointer: hovered, opened on a click. Titled nodes handle both themselves.
const HIT_RADIUS = 8;
function dotAt(event: MouseEvent) {
  if ((event.target as HTMLElement).closest('a')) return null;
  const rect = view.value!.getBoundingClientRect();
  const x = event.clientX - rect.left;
  const y = event.clientY - rect.top;
  let best: string | null = null;
  let bestDistance = HIT_RADIUS;
  for (const id of visible.value) {
    if (titled.value.has(id)) continue;
    const at = px(id);
    const distance = Math.hypot(at.x - x, at.y - y);
    if (distance < bestDistance) [best, bestDistance] = [id, distance];
  }
  return best;
}

function hoverDot(event: PointerEvent) {
  if (drag || (event.target as HTMLElement).closest('a')) return;
  const id = dotAt(event);
  if (id) hover(id);
  else if (hovered.value && !titled.value.has(hovered.value)) hovered.value = null;
}

function openDot(event: MouseEvent) {
  const id = dotAt(event);
  if (id) navigateTo(isTag(id) ? tagPath(id.slice(1)) : `/notes/${id}`);
}
</script>

<template>
  <div
    ref="view"
    class="relative size-full cursor-grab touch-none overflow-hidden select-none active:cursor-grabbing"
    :class="{ 'rounded-md bg-muted': mini }"
    :data-settled="settled || undefined"
    @pointerdown="startDrag"
    @pointermove="hoverDot"
    @pointerleave="hovered = null"
    @click.capture="swallowClickAfterDrag"
    @click="openDot"
  >
    <div v-if="!settled" class="absolute inset-0 flex items-center justify-center" role="status">
      <UIcon
        :name="appConfig.ui.icons.loading"
        class="size-5 animate-spin text-muted"
        :aria-label="$t('graph.loading')"
      />
    </div>
    <canvas ref="canvas" class="pointer-events-none absolute inset-0 size-full" aria-hidden="true" />
    <div class="absolute inset-0" :style="{ transform: `translate(${offset.x}px, ${offset.y}px)` }">
      <NuxtLink
        v-for="note in titledNotes"
        :key="note.id"
        :to="`/notes/${note.id}`"
        draggable="false"
        :style="{ left: `${at(note.id).x}%`, top: `${at(note.id).y}%` }"
        :class="[{ 'z-10': focus.has(note.id) }, mini ? '-translate-y-1/2' : '-translate-y-[5px]']"
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
        v-for="tag in titledTags"
        :key="tagNode(tag)"
        :to="tagPath(tag)"
        draggable="false"
        :style="{ left: `${at(tagNode(tag)).x}%`, top: `${at(tagNode(tag)).y}%` }"
        :class="{ 'z-10': focus.has(tagNode(tag)) }"
        :aria-label="mini ? `#${tag}` : undefined"
        class="absolute max-w-32 -translate-x-1/2 -translate-y-1/2 rounded-sm px-0.5"
        @mouseenter="hover(tagNode(tag))"
        @mouseleave="hovered = null"
        @focus="hover(tagNode(tag))"
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
          :disabled="zoom >= maxZoom"
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
