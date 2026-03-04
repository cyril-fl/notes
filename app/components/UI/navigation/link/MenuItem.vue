<script lang="ts" setup>
import type { NavigationMenuItem } from '~/types/ui';

const props = defineProps<{
  item: NavigationMenuItem;
}>();

const dropRef = useTemplateRef('dropRef');
const isNotesLink = computed(() => props.item.id === 'notes');
const rootFolderId = ref<string | null>(null);
const { isOver } = useDropZone({
  elementRef: isNotesLink.value ? dropRef : ref(null),
  folderId: rootFolderId,
});
</script>

<template>
  <li ref="dropRef" class="transition-colors" :class="{ 'bg-primary/10 rounded-md': isOver }">
    <UButton
      :icon="item.icon"
      :label="$t(item.label)"
      :to="item.to"
      class="w-full justify-start select-none"
      color="neutral"
      variant="ghost"
    />
  </li>
</template>
