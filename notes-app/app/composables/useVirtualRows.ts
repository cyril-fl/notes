import { useVirtualizer } from '@tanstack/vue-virtual';
import type { ShallowRef, WatchSource } from 'vue';

/**
 * Only the rows of a list or a mosaic in sight of the panel's scroll, the others standing as padding: a list may hold
 * 50,000 notes. row is the height of a list row, caption what a tile shows under its square; moved changes when
 * something above the list grows or shrinks. Gives the items to render, from start to end excluded.
 */
export function useVirtualRows(
  list: Readonly<ShallowRef<HTMLElement | null>>,
  count: () => number,
  grid: () => boolean,
  { row, caption }: { row: number; caption: number },
  moved: WatchSource = () => 0,
) {
  const { width } = useElementSize(list);
  // As .mosaic lays them out: tiles of 7.5rem at least, 1rem apart.
  const columns = computed(() => (grid() ? Math.max(1, Math.floor((width.value + 16) / 136)) : 1));
  // A tile is square, then gap-2, its caption and gap-y-5.
  const rowHeight = computed(() =>
    grid() ? (width.value - (columns.value - 1) * 16) / columns.value + 8 + caption + 20 : row,
  );
  // Where the list starts in the panel's scroll, below the header and whatever sits above it.
  const scrollMargin = ref(0);
  const scroller = () => list.value?.closest<HTMLElement>('[data-slot="body"]') ?? null;
  watch(
    [list, moved],
    () => {
      const element = scroller();
      if (list.value && element)
        scrollMargin.value =
          list.value.getBoundingClientRect().top - element.getBoundingClientRect().top + element.scrollTop;
    },
    { flush: 'post' },
  );
  const virtualizer = useVirtualizer(
    computed(() => {
      const size = rowHeight.value;
      return {
        count: Math.ceil(count() / columns.value),
        getScrollElement: scroller,
        estimateSize: () => size,
        scrollMargin: scrollMargin.value,
        overscan: 5,
      };
    }),
  );
  watch(rowHeight, () => virtualizer.value.measure());
  return computed(() => {
    const rows = virtualizer.value.getVirtualItems();
    const first = rows[0];
    const last = rows.at(-1);
    if (!first || !last) return { start: 0, end: 0, before: 0, after: 0 };
    return {
      start: first.index * columns.value,
      end: Math.min(count(), (last.index + 1) * columns.value),
      before: first.start - scrollMargin.value,
      after: virtualizer.value.getTotalSize() - (last.end - scrollMargin.value),
    };
  });
}
