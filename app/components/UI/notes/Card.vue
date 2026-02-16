<script setup lang="ts">
import type { ContextMenuItem } from '@nuxt/ui';

// Define
interface NotesCardProps {
  item: Note;
}

const props = defineProps<NotesCardProps>();
const { deleteItem } = useActions();

// Data
const actions = ref<ContextMenuItem[][]>([
  [
    // CRUD
    deleteItem(props.item.id),
  ],
]);
// Methods

// Lifecycle

// SEO
</script>

<template>
  <UContextMenu :items="actions" class="flex flex-col gap-2">
    <NuxtLink
      :to="`${NAVIGATION.notes}${item.id}`"
      as="p"
      class="bg-muted text-xs text-default p-4 rounded-md aspect-square size-30 overflow-hidden"
    >
      <span class="line-clamp-5">
        {{ item.content }}
      </span>
    </NuxtLink>
    <ul class="text-xs text-muted text-center">
      <li>{{ item.updatedAt }}</li>
    </ul>
  </UContextMenu>
</template>
