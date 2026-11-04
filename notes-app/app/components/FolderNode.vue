<script setup lang="ts">
import type { Item } from '~/types/items';

const props = defineProps<{ folder: Item }>();

const route = useRoute();
const { itemMenu } = useItems();

const open = useState(`open:${props.folder.id}`, () => false);

// The whole node drags, so the check of useDropFolder sees the subfolders inside it.
const node = useTemplateRef('node');
const row = useTemplateRef('row');
useDragItem(node, () => props.folder);
const isOver = useDropFolder(row, () => props.folder.id);

const { start, stop } = useTimeoutFn(() => (open.value = true), 600, { immediate: false });
watch(isOver, (over) => (over ? start() : stop()));
</script>

<template>
  <div ref="node">
    <UContextMenu :items="itemMenu(folder)">
      <div
        ref="row"
        class="row pl-0.5"
        :class="{ 'row-active': route.params.id === folder.id, 'row-over': isOver }"
      >
        <button
          type="button"
          class="flex size-5 shrink-0 items-center justify-center rounded"
          :aria-label="open ? 'Replier' : 'Déplier'"
          :aria-expanded="open"
          @click="open = !open"
        >
          <UIcon
            name="i-lucide-chevron-right"
            class="size-3.5 text-muted transition-transform duration-150"
            :class="{ 'rotate-90': open }"
          />
        </button>
        <NuxtLink :to="`/folders/${folder.id}`" draggable="false" class="flex min-w-0 flex-1 items-center gap-2 self-stretch">
          <UIcon name="i-lucide-folder" class="size-4 shrink-0" :class="isOver ? 'text-white' : 'text-primary'" />
          <span class="truncate">{{ folder.title }}</span>
        </NuxtLink>
      </div>
    </UContextMenu>
    <FolderTree v-if="open" :parent-id="folder.id" />
  </div>
</template>
