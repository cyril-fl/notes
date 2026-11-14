<script setup lang="ts">
import type { SavedSearch } from '~/types/items';

const { icons } = useIcons();
const props = defineProps<{ search: SavedSearch }>();

const { editSavedSearch, removeSavedSearch, reorderSavedSearch } = useSavedSearches();

const row = useTemplateRef('row');
useReorderItem(row, () => props.search.id);
const isOver = useDropReorder(row, () => props.search.id, reorderSavedSearch);
</script>

<template>
  <li ref="row" class="flex items-center gap-3 px-4 py-2" :class="{ 'bg-elevated': isOver }">
    <UIcon :name="icons.savedSearch" class="size-4 shrink-0 text-muted" />
    <p class="min-w-0 flex-1 truncate text-sm">{{ search.name }}</p>
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
