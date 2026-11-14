<script setup lang="ts">
import { useVirtualizer } from '@tanstack/vue-virtual';
import type { Item } from '~/types/items';

const { icons } = useIcons();

const props = defineProps<{ folderId?: string; title: string; subtitle?: string }>();

const { createFolder, moveItem, removeItems, exportItems } = useItems();
const view = useView();
const { items, status, load } = useFolderPages(props.folderId);

// ⌘-click (Ctrl elsewhere) adds or takes out an item, ⇧-click selects a range from the last one; a plain click opens it.
const selection = useSelection();
const selectedIds = computed(() => new Set(selection.value));
const selected = computed(() => items.value.filter((item) => item && selectedIds.value.has(item.id)) as Item[]);
let anchor = -1;

function select(event: MouseEvent, index: number) {
  const item = items.value[index];
  if (!item || (!event.metaKey && !event.ctrlKey && !event.shiftKey)) return;
  event.preventDefault();
  event.stopPropagation(); // The link under it must not open.
  if (event.shiftKey && anchor >= 0) {
    // Only the items read so far: the range skips the pages not read yet.
    selection.value = selectRange(
      selection.value,
      items.value.map((item) => item?.id),
      anchor,
      index,
    );
    return;
  }
  selection.value = toggleSelected(selection.value, item.id);
  anchor = index;
}

onKeyStroke('Escape', () => (selection.value = []));

// A selected folder cannot go into itself: it is not offered.
const destinations = useDestinations(() => selection.value);

async function moveSelection(destination: string) {
  await moveItem(selection.value, destination === ROOT ? null : destination);
  selection.value = [];
}

async function removeSelection() {
  await removeItems(selected.value);
  selection.value = [];
}

const panel = useTemplateRef('panel');
const isOver = useDropFolder(panel, () => props.folderId ?? null);

// Only the rows in sight are rendered, the others stand as padding: a folder may hold 50,000 notes.
const list = useTemplateRef('list');
const { width } = useElementSize(list);
// As .mosaic lays them out: tiles of 7.5rem at least, 1rem apart.
const columns = computed(() => (view.value === 'grid' ? Math.max(1, Math.floor((width.value + 16) / 136)) : 1));
// A row is h-9; a tile is square, then gap-2, two text-xs lines and gap-y-5.
const rowHeight = computed(() =>
  view.value === 'grid' ? (width.value - (columns.value - 1) * 16) / columns.value + 8 + 32 + 20 : 36,
);
// Where the list starts in the panel's scroll, below the header and the selection bar.
const scrollMargin = ref(0);
const scroller = () => panel.value?.closest<HTMLElement>('[data-slot="body"]') ?? null;
watch(
  [list, () => selected.value.length > 0],
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
      count: Math.ceil(items.value.length / columns.value),
      getScrollElement: scroller,
      estimateSize: () => size,
      scrollMargin: scrollMargin.value,
      overscan: 5,
    };
  }),
);
watch(rowHeight, () => virtualizer.value.measure());
const visible = computed(() => {
  const rows = virtualizer.value.getVirtualItems();
  const first = rows[0];
  const last = rows.at(-1);
  if (!first || !last) return { start: 0, items: [] as (Item | undefined)[], before: 0, after: 0 };
  const start = first.index * columns.value;
  return {
    start,
    items: items.value.slice(start, (last.index + 1) * columns.value),
    before: first.start - scrollMargin.value,
    after: virtualizer.value.getTotalSize() - (last.end - scrollMargin.value),
  };
});
// The rows in sight that are not read yet come by pages.
watchEffect(() => load(visible.value.start, visible.value.start + visible.value.items.length - 1));
</script>

<template>
  <div
    ref="panel"
    class="flex min-h-full flex-col p-8 transition-shadow duration-150"
    :class="{ 'ring-2 ring-inset ring-primary/40': isOver }"
  >
    <PageHeader :title="title">
      <ViewToggle />
      <UTooltip :text="$t('common.newFolder')">
        <UButton
          :icon="icons.folderAdd"
          color="neutral"
          variant="ghost"
          :aria-label="$t('common.newFolder')"
          @click="createFolder(folderId)"
        />
      </UTooltip>
      <UTooltip :text="$t('common.newNote')">
        <UButton
          :icon="icons.noteAdd"
          color="neutral"
          variant="ghost"
          :aria-label="$t('common.newNote')"
          :to="newNotePath(folderId)"
        />
      </UTooltip>
    </PageHeader>
    <div
      v-if="selected.length"
      class="sticky top-0 z-10 mb-3 flex items-center gap-2 rounded-md bg-elevated py-1 pr-1 pl-3 text-sm"
    >
      <span class="mr-auto">{{ $t('folder.selected', selected.length) }}</span>
      <USelectMenu
        :items="destinations"
        value-key="value"
        :placeholder="$t('folder.moveToPlaceholder')"
        size="sm"
        class="w-56"
        :aria-label="$t('folder.moveTo')"
        @update:model-value="moveSelection"
      />
      <UButton
        :label="$t('settings.export')"
        :icon="icons.export"
        color="neutral"
        variant="ghost"
        size="sm"
        @click="exportItems(selection)"
      />
      <UButton
        :label="$t('common.delete')"
        :icon="icons.delete"
        color="error"
        variant="ghost"
        size="sm"
        @click="removeSelection"
      />
      <UTooltip :text="$t('folder.clearSelection')" :kbds="['esc']">
        <UButton
          :icon="icons.close"
          color="neutral"
          variant="ghost"
          size="sm"
          :aria-label="$t('folder.clearSelection')"
          @click="selection = []"
        />
      </UTooltip>
    </div>
    <ul
      v-if="items.length"
      ref="list"
      :class="{ mosaic: view === 'grid' }"
      :style="{ paddingTop: `${visible.before}px`, paddingBottom: `${visible.after}px` }"
    >
      <li
        v-for="(item, index) in visible.items"
        :key="item?.id ?? visible.start + index"
        @click.capture="select($event, visible.start + index)"
      >
        <ItemRow v-if="item" :item="item" :grid="view === 'grid'" :selected="selectedIds.has(item.id)" />
        <!-- A tile is a square, then gap-2 and its two lines of text. -->
        <USkeleton v-else :class="view === 'grid' ? 'mb-10 aspect-square' : 'h-9'" />
      </li>
    </ul>
    <UEmpty
      v-else-if="status === 'success'"
      :icon="icons.folderEmpty"
      :title="$t('folder.empty')"
      :description="$t('folder.emptyDescription')"
      :actions="[
        { label: $t('common.newNote'), icon: icons.noteAdd, to: newNotePath(folderId) },
        {
          label: $t('common.newFolder'),
          icon: icons.folderAdd,
          color: 'neutral',
          variant: 'subtle',
          onClick: () => createFolder(folderId),
        },
      ]"
    />
    <!-- Like a note's footer. The subtitle is the folder's own path: the breadcrumb leads to its parents. -->
    <footer v-if="subtitle" class="sticky bottom-0 z-10 -mx-8 -mb-8 mt-auto border-t border-default bg-default px-8">
      <div class="flex h-9 items-center text-xs text-muted">
        <PathNav :path="subtitle.slice(0, subtitle.lastIndexOf('/'))" />
      </div>
    </footer>
  </div>
</template>
