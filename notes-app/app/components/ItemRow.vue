<script setup lang="ts">
import type { Item } from '~/types/items';

const { icons } = useIcons();
const { locale } = useI18n();

const props = defineProps<{ item: Item; grid?: boolean; selected?: boolean }>();

const { itemMenu } = useItems();

const link = useTemplateRef('link');
useDragItem(link, () => props.item);

const isFolder = computed(() => props.item.kind === 'folder');
const to = computed(() => `/${isFolder.value ? 'folders' : 'notes'}/${props.item.id}`);
// Only folders take drops: a note target always refuses.
const isOver = useDropFolder(
  computed(() => (isFolder.value ? link.value : null)),
  () => props.item.id,
);
const date = computed(() =>
  new Date(props.item.updated_at ?? props.item.created_at).toLocaleDateString(locale.value, {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  }),
);
</script>

<template>
  <UContextMenu :items="itemMenu(item)">
    <NuxtLink v-if="grid" ref="link" :to="to" class="group flex min-w-0 flex-col gap-2">
      <div class="tile" :class="{ 'tile-over': isOver, 'tile-selected': selected && !isOver }">
        <UIcon :name="isFolder ? icons.folder : icons.note" class="size-10 text-dimmed" />
      </div>
      <div class="min-w-0 text-center text-xs">
        <p class="truncate">{{ displayTitle(item.title, $t('untitled')) }}</p>
        <p class="text-muted">{{ date }}</p>
      </div>
    </NuxtLink>
    <NuxtLink
      v-else
      ref="link"
      :to="to"
      class="row h-9"
      :class="{ 'row-over': isOver, 'row-selected': selected && !isOver }"
    >
      <UIcon
        :name="isFolder ? icons.folder : icons.note"
        class="size-4 shrink-0"
        :class="isOver ? 'text-inverted' : isFolder ? 'text-primary' : 'text-muted'"
      />
      <span class="truncate">{{ displayTitle(item.title, $t('untitled')) }}</span>
      <span class="ml-auto shrink-0 pl-4 text-xs" :class="isOver ? 'text-inverted' : 'text-muted'">{{ date }}</span>
    </NuxtLink>
  </UContextMenu>
</template>
