<script setup lang="ts">
import type { SavedSearch } from '~/types/items';

const { icons } = useIcons();
const props = defineProps<{ search: SavedSearch }>();
const route = useRoute();

const { savedSearchMenu, reorderSavedSearch } = useSavedSearches();

const link = useTemplateRef('link');
useReorderItem(link, () => props.search.id);
const isOver = useDropReorder(link, () => props.search.id, reorderSavedSearch);
</script>

<template>
  <UContextMenu :items="savedSearchMenu(search)">
    <NuxtLink
      ref="link"
      :to="`/searches/${search.id}`"
      class="row"
      :class="{ 'row-active': route.params.id === search.id, 'row-over': isOver }"
    >
      <UIcon :name="icons.search" class="size-4 shrink-0" :class="isOver ? 'text-inverted' : 'text-muted'" />
      <span class="truncate">{{ search.name }}</span>
    </NuxtLink>
  </UContextMenu>
</template>
