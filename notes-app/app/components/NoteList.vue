<script setup lang="ts">
type Note = { id: string; title: string; path: string; snippet?: string; readonly?: boolean };

const props = defineProps<{
  // A hole stands for a note of a page not read yet.
  notes: (Note | undefined)[];
  grid?: boolean;
  /** Reads the notes from one index to another: a list read by pages renders only the rows in sight. */
  load?: (from: number, to: number) => void;
}>();

const { icons } = useIcons();

// A row is py-2, a text-sm line and a text-xs one; a tile has one text-xs line under its square.
const list = useTemplateRef('list');
const rows = useVirtualRows(
  list,
  () => (props.load ? props.notes.length : 0),
  () => !!props.grid,
  { row: 52, caption: 16 },
);
const visible = computed(() => {
  // Not paged, as the related notes in their own scroll box: short, rendered whole.
  if (!props.load) return { start: 0, items: props.notes, before: 0, after: 0 };
  return { ...rows.value, items: props.notes.slice(rows.value.start, rows.value.end) };
});
watchEffect(() => props.load?.(visible.value.start, visible.value.start + visible.value.items.length - 1));
</script>

<template>
  <ul
    ref="list"
    :class="{ mosaic: grid }"
    :style="{ paddingTop: `${visible.before}px`, paddingBottom: `${visible.after}px` }"
  >
    <li v-for="(note, index) in visible.items" :key="note?.id ?? visible.start + index" class="min-w-0">
      <template v-if="!note">
        <!-- A tile is a square, then gap-2 and its line of text. -->
        <USkeleton :class="grid ? 'mb-6 aspect-square' : 'h-13'" />
      </template>
      <NuxtLink v-else-if="grid" :to="`/notes/${note.id}`" class="group flex flex-col gap-2">
        <div class="tile relative items-start justify-start text-muted">
          <p class="line-clamp-6">{{ note.snippet ?? note.path }}</p>
          <UIcon
            v-if="note.readonly"
            :name="icons.lock"
            :aria-label="$t('note.readonly')"
            class="absolute top-2 right-2 size-4 text-muted"
          />
        </div>
        <p class="truncate text-center text-xs">{{ displayTitle(note.title, $t('untitled')) }}</p>
      </NuxtLink>
      <NuxtLink
        v-else
        :to="`/notes/${note.id}`"
        class="block rounded-md px-2 py-2 transition-colors duration-150 hover:bg-accented/60"
      >
        <p class="flex items-center gap-2 text-sm">
          <span class="truncate">{{ displayTitle(note.title, $t('untitled')) }}</span>
          <UIcon
            v-if="note.readonly"
            :name="icons.lock"
            :aria-label="$t('note.readonly')"
            class="ml-auto size-3.5 shrink-0 text-muted"
          />
        </p>
        <p class="truncate text-xs text-muted">{{ note.snippet ?? note.path }}</p>
      </NuxtLink>
    </li>
  </ul>
</template>
