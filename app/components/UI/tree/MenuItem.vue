<script setup lang="ts">
import {
  EditableArea,
  EditableInput,
  EditablePreview,
  EditableRoot,
} from 'reka-ui';

import type { ContextMenuItem } from '@nuxt/ui';
import type { NavigationTreeProps } from '~/types/ui';

// Define
const { $hooks } = useNuxtApp();
const props = defineProps<NavigationTreeProps>();

const model = ref(props.item.label);
const editable = useTemplateRef('editable');
const inputRef = useTemplateRef('inputRef');

const { addFolder, addNote, deleteItem, updateFolderLink } = useActions();

// Data
const actions = ref<ContextMenuItem[][]>([
  [
    // ADD
    addFolder(props.item.id, (newId) => {
      $hooks.callHook('folder-navigation:blur-except', props.item.id);
      $hooks.callHook('folder-navigation:update', newId);
    }),
    addNote(props.item.id),
  ],
  [
    // EDIT
    updateFolderLink(props.item.id),
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
  if (!trimmedValue || trimmedValue === props.item.label) return;

  $hooks.callHook('data:update', props.item.id, { title: trimmedValue });
  $hooks.callHook('card:folder:refresh', props.item.id, trimmedValue);
}

function handleBlur(id: string) {
  if (id !== props.item.id) return;
  editable.value?.submit();
}

function handleBlurExcept(id: string) {
  if (id === props.item.id) return;
  editable.value?.submit();
}

// Lifecycle
watch(
  () => props.item,
  (newLabel) => {
    model.value = newLabel.label;
  }
);

onMounted(() => {
  $hooks.addHooks({
    'folder-navigation:update': handleEdit,
    'folder-navigation:blur': handleBlur,
    'folder-navigation:blur-except': handleBlurExcept,
  });
});
// SEO
</script>

<template>
  <li>
    <UContextMenu :items="actions">
      <UButton
        :icon="item.icon"
        :to="item.to"
        class="w-full justify-start"
        color="neutral"
        variant="ghost"
      >
        <EditableRoot
          ref="editable"
          v-slot="{ isEditing, submit }"
          v-model="model"
          placeholder="Enter text..."
          activation-mode="none"
          auto-resize
          @submit="handleSubmit"
        >
          <EditableArea
            :class="[isEditing ? 'bg-accented px-1 py-0.5' : '']"
            class="rounded-md"
            @keydown.enter.prevent="submit"
          >
            <EditablePreview />
            <EditableInput as-child>
              <input ref="inputRef" type="text" @click.stop />
            </EditableInput>
          </EditableArea>
        </EditableRoot>
      </UButton>
    </UContextMenu>
  </li>
</template>
