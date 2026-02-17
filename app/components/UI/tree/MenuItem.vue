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
const { state: folderState, requestEdit, clearEdit } = useFolder();

const model = ref(props.item.label);
const editable = useTemplateRef('editable');
const inputRef = useTemplateRef('inputRef');

const { addFolder, addNote, deleteItem } = useActions();

// Data

/** Ce MenuItem est-il la cible de l'édition inline (rename) ? */
const isEditTarget = computed(
  () => folderState.value.editingId === props.item.id
);

const actions = ref<ContextMenuItem[][]>([
  [
    // ADD — ouvre la modale de création
    addFolder(props.item.id),
    addNote(props.item.id),
  ],
  [
    // EDIT — rename inline via Editable
    {
      label: 'Renommer',
      icon: 'i-lucide-pencil',
      onSelect: () => requestEdit(props.item.id),
      disabled: props.item.id === 'root',
    },
    deleteItem(props.item.id),
  ],
]);

// Methods

function handleSubmit(value: string | null | undefined) {
  clearEdit(props.item.id);

  if (!value) return;
  const trimmedValue = value.trim();
  if (!trimmedValue || trimmedValue === props.item.label) return;

  $hooks.callHook('data:update', props.item.id, { title: trimmedValue });
  $hooks.callHook('card:folder:refresh', props.item.id, trimmedValue);
}

// Lifecycle
watch(
  () => props.item.label,
  (newLabel) => {
    model.value = newLabel;
  }
);

/** Active le mode édition inline lorsque requestEdit est appelé */
watch(isEditTarget, async (isTarget) => {
  if (!isTarget) return;

  editable.value?.edit();
  await nextTick();

  const input = inputRef.value;
  if (!input) return;
  input.focus({ preventScroll: true });
  const end = input.value.length;
  input.setSelectionRange(end, end);
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
