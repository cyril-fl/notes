<script setup lang="ts">
import type { ContextMenuItem } from '@nuxt/ui';
import {
  EditableArea,
  EditableInput,
  EditablePreview,
  EditableRoot,
} from 'reka-ui';

// Define
interface FolderCardProps {
  item: Folder;
}

const { $hooks } = useNuxtApp();
const { icons } = useIcons();
const props = defineProps<FolderCardProps>();

const model = ref(props.item.title);
const editable = useTemplateRef('editable');
const inputRef = useTemplateRef('inputRef');
const hookDisposers: Array<() => void> = [];
const { getChildrenCountByType } = useDataUtils();

const { addNote, deleteItem, updateFolderCard } = useActions();

// Data
const notesChild = computed(() =>
  getChildrenCountByType(props.item.id, { types: ItemType.NOTE })
);

const actions = ref<ContextMenuItem[][]>([
  [
    // CRUD Note
    addNote(props.item.id),
  ],
  [
    // CRUD Folder
    updateFolderCard(props.item.id),
    deleteItem(props.item.id),
  ],
]);

// Methods
async function handleEdit(id: string) {
  if (id !== props.item.id) return;
  editable.value?.edit();
  await nextTick();
  setTimeout(() => {
    const input = inputRef.value;
    if (input) {
      input.focus();
      input.selectionStart = input.selectionEnd = input.value.length;
    }
  }, 50);
}

function handleSubmit(value: string | null | undefined) {
  if (!value) return;
  const trimmedValue = value.trim();
  if (!trimmedValue || trimmedValue === props.item.title) return;
  $hooks.callHook('data:update', props.item.id, { title: trimmedValue });
  $hooks.callHook('folder-navigation:refresh', props.item.id, trimmedValue);
}

// Lifecycle
watch(
  () => props.item.title,
  (newTitle) => {
    model.value = newTitle;
  }
);

onMounted(() => {
  hookDisposers.push($hooks.hook('folder-card:update', handleEdit));
});

onBeforeUnmount(() => {
  hookDisposers.forEach((dispose) => dispose());
});
// SEO
</script>

<template>
  <UContextMenu
    :items="actions"
    class="flex flex-col gap-2 max-w-30 overflow-hidden"
  >
    <NuxtLink
      :to="`/${item.id}`"
      as="div"
      class="bg-muted text-xs text-default p-4 rounded-md aspect-square size-30 overflow-hidden flex items-center justify-center"
    >
      <UIcon :name="icons.folder" class="size-10" />
    </NuxtLink>
    <ul class="text-xs text-muted text-center truncate text-ellipsis">
      <EditableRoot
        ref="editable"
        v-slot="{ submit }"
        v-model="model"
        placeholder="Enter text..."
        activation-mode="none"
        as="li"
        @submit="handleSubmit"
      >
        <EditableArea class="rounded-md" @keydown.enter.prevent="submit">
          <EditablePreview />
          <EditableInput as-child>
            <input
              ref="inputRef"
              type="text"
              class="bg-transparent p-0 max-w-30 text-xs"
              @click.stop
            />
          </EditableInput>
        </EditableArea>
      </EditableRoot>
      <li>{{ $t('pages.folder.notes_count', { count: notesChild }) }}</li>
    </ul>
  </UContextMenu>
</template>

<style scoped></style>
