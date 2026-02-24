<script setup lang="ts">
import type { ContextMenuItem } from '@nuxt/ui';
import {
  EditableArea,
  EditableInput,
  EditablePreview,
  EditableRoot,
} from 'reka-ui';

interface FolderCardProps {
  item: Folder;
}

const props = defineProps<FolderCardProps>();
const { update } = useDataActions();
const { onFolderCardUpdate } = useUIEvents();
const { getChildrenCountByType } = useDataUtils();

const itemIdRef = computed(() => props.item.id);
const folderActions = useActions(itemIdRef, { requestEdit: () => {} });

const model = ref(props.item.title);
const editable = useTemplateRef('editable');
const inputRef = useTemplateRef('inputRef');

const { icons } = useIcons();

const notesChild = computed(() =>
  getChildrenCountByType(props.item.id, { types: ItemType.NOTE })
);

const actions = computed<ContextMenuItem[][]>(() => [
  [folderActions.addNote()],
  [folderActions.updateFolderCard(), folderActions.deleteItem()],
]);

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
  update(props.item.id, { title: trimmedValue });
}

watch(
  () => props.item.title,
  (newTitle) => {
    model.value = newTitle;
  }
);

onMounted(() => {
  const dispose = onFolderCardUpdate(handleEdit);
  onBeforeUnmount(dispose);
});
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
    <ul class="">
      <EditableRoot
        ref="editable"
        v-slot="{ submit }"
        v-model="model"
        placeholder="Enter text..."
        activation-mode="none"
        as="li"
        @submit="handleSubmit"
      >
        <EditableArea
          class="rounded-md text-xs text-muted text-center truncate text-ellipsis"
          @keydown.enter.prevent="submit"
        >
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
