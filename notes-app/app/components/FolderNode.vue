<script setup lang="ts">
import type { Item } from '~/types/items';

const { icons } = useIcons();

const props = defineProps<{ folder: Item }>();

const route = useRoute();
const { itemMenu } = useItems();

const open = useState(`open:${props.folder.id}`, () => false);

// shortcut: reads each shown folder's children one level ahead to know if it has subfolders (one request per folder); have the API tell it if that grows slow.
const { data: children } = useChildren(props.folder.id);
const hasSubfolders = computed(() => children.value.some((item) => item.kind === 'folder'));

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
        class="row"
        :class="{ 'row-active': route.params.id === folder.id, 'row-over': isOver }"
      >
        <NuxtLink :to="`/folders/${folder.id}`" draggable="false" class="flex min-w-0 flex-1 items-center gap-2 self-stretch">
          <UIcon :name="icons.folder" class="size-4 shrink-0" :class="isOver ? 'text-inverted' : 'text-muted'" />
          <span class="truncate">{{ folder.title }}</span>
        </NuxtLink>
        <button
          v-if="hasSubfolders"
          type="button"
          class="flex size-5 shrink-0 items-center justify-center rounded"
          :aria-label="open ? $t('items.collapse') : $t('items.expand')"
          :aria-expanded="open"
          @click="open = !open"
        >
          <UIcon
            :name="icons.chevronRight"
            class="size-3.5 text-muted transition-transform duration-150"
            :class="{ 'rotate-90': open }"
          />
        </button>
      </div>
    </UContextMenu>
    <FolderTree v-if="open" :parent-id="folder.id" />
  </div>
</template>
