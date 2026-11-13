<script setup lang="ts">
import type { SavedSearch } from '~/types/items';

const { icons } = useIcons();
const props = defineProps<{ search: SavedSearch }>();

const { editSavedSearch, removeSavedSearch, moveSavedSearch, reorderSavedSearch } = useSavedSearches();

const row = useTemplateRef('row');
useReorderItem(row, () => props.search.id);
const isOver = useDropReorder(row, () => props.search.id, reorderSavedSearch);
</script>

<template>
  <li ref="row" class="flex items-center gap-3 px-4 py-2" :class="{ 'bg-elevated': isOver }">
    <UIcon :name="icons.search" class="size-4 shrink-0 text-muted" />
    <p class="min-w-0 flex-1 truncate text-sm">{{ search.name }}</p>
    <UButton
      :icon="icons.up"
      color="neutral"
      variant="ghost"
      size="xs"
      :aria-label="$t('search.saved.moveUp')"
      @click="moveSavedSearch(search.id, 'up')"
    />
    <UButton
      :icon="icons.down"
      color="neutral"
      variant="ghost"
      size="xs"
      :aria-label="$t('search.saved.moveDown')"
      @click="moveSavedSearch(search.id, 'down')"
    />
    <UButton
      :icon="icons.rename"
      color="neutral"
      variant="ghost"
      :aria-label="$t('common.edit')"
      @click="editSavedSearch(search)"
    />
    <UButton
      :icon="icons.delete"
      color="neutral"
      variant="ghost"
      :aria-label="$t('common.delete')"
      @click="removeSavedSearch(search)"
    />
  </li>
</template>
